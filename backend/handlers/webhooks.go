package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"webhook-vault/backend/database"
	"webhook-vault/backend/middleware"
	"webhook-vault/backend/models"
)

// IngestWebhook captures external HTTP webhooks
func IngestWebhook(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		middleware.JSONError(w, "missing endpoint slug in path", http.StatusBadRequest)
		return
	}

	var endpoint models.Endpoint
	err := database.DB.QueryRow(
		`SELECT id, slug, name, target_url, secret_token, is_active, total_received, total_forwarded, created_at 
		 FROM endpoints WHERE slug = ?`, slug,
	).Scan(&endpoint.ID, &endpoint.Slug, &endpoint.Name, &endpoint.TargetURL,
		&endpoint.SecretToken, &endpoint.IsActive, &endpoint.TotalReceived,
		&endpoint.TotalForwarded, &endpoint.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.JSONError(w, fmt.Sprintf("endpoint '%s' not found", slug), http.StatusNotFound)
			return
		}
		middleware.JSONError(w, "database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !endpoint.IsActive {
		middleware.JSONError(w, fmt.Sprintf("endpoint '%s' is inactive", slug), http.StatusServiceUnavailable)
		return
	}

	// Secret token check if specified
	if endpoint.SecretToken != "" {
		sig := r.Header.Get("X-Vault-Signature")
		if sig == "" {
			sig = r.Header.Get("X-Hub-Signature-256")
		}
		if sig == "" {
			sig = r.Header.Get("Stripe-Signature")
		}
		// If custom header configured and present, verify match
		if reqToken := r.Header.Get("X-Webhook-Secret"); reqToken != "" && reqToken != endpoint.SecretToken {
			middleware.JSONError(w, "invalid signature secret", http.StatusUnauthorized)
			return
		}
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 5*1024*1024)) // 5MB cap
	if err != nil {
		middleware.JSONError(w, "failed to read body payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Capture headers
	headerMap := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headerMap[k] = v[0]
		}
	}
	headersJSON, _ := json.Marshal(headerMap)

	// Capture Query params
	queryMap := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			queryMap[k] = v[0]
		}
	}
	queryJSON, _ := json.Marshal(queryMap)

	reqID := fmt.Sprintf("REQ-%s-%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
	cType := r.Header.Get("Content-Type")
	if cType == "" {
		cType = "application/json"
	}
	srcIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		srcIP = strings.Split(forwarded, ",")[0]
	}

	res, err := database.DB.Exec(
		`INSERT INTO webhook_events (endpoint_id, request_id, method, path, headers, query_params, body, content_type, source_ip, byte_size, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'RECEIVED')`,
		endpoint.ID, reqID, r.Method, r.URL.Path, string(headersJSON), string(queryJSON),
		string(bodyBytes), cType, srcIP, len(bodyBytes),
	)
	if err != nil {
		middleware.JSONError(w, "failed to store webhook event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	eventID, _ := res.LastInsertId()
	_, _ = database.DB.Exec("UPDATE endpoints SET total_received = total_received + 1 WHERE id = ?", endpoint.ID)

	// Dispatch to target URL if configured
	var attempt *models.DeliveryAttempt
	finalStatus := "RECEIVED"

	if endpoint.TargetURL != "" {
		att := executeDispatch(eventID, 1, endpoint.TargetURL, r.Method, bodyBytes, headerMap)
		attempt = &att
		if att.HTTPStatus >= 200 && att.HTTPStatus < 300 {
			finalStatus = "FORWARDED"
			_, _ = database.DB.Exec("UPDATE endpoints SET total_forwarded = total_forwarded + 1 WHERE id = ?", endpoint.ID)
		} else {
			finalStatus = "FAILED"
		}
		_, _ = database.DB.Exec("UPDATE webhook_events SET status = ? WHERE id = ?", finalStatus, eventID)
	}

	middleware.JSONSuccess(w, http.StatusAccepted, map[string]any{
		"request_id":       reqID,
		"event_id":         eventID,
		"endpoint":         endpoint.Slug,
		"status":           finalStatus,
		"delivery_attempt": attempt,
	}, "webhook captured and logged")
}

// ListEvents returns paginated and filtered events
func ListEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	epID := strings.TrimSpace(q.Get("endpoint_id"))
	status := strings.TrimSpace(q.Get("status"))
	search := strings.TrimSpace(q.Get("search"))

	query := `SELECT e.id, e.endpoint_id, e.request_id, e.method, e.path, e.headers, e.query_params,
	                 e.body, e.content_type, e.source_ip, e.byte_size, e.status, e.created_at,
	                 ep.slug, ep.name, ep.target_url
	          FROM webhook_events e
	          JOIN endpoints ep ON e.endpoint_id = ep.id
	          WHERE 1=1`
	var args []any

	if epID != "" && epID != "0" && epID != "ALL" {
		query += " AND e.endpoint_id = ?"
		args = append(args, epID)
	}

	if status != "" && status != "ALL" {
		query += " AND e.status = ?"
		args = append(args, status)
	}

	if search != "" {
		query += " AND (e.request_id LIKE ? OR e.body LIKE ? OR ep.name LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}

	query += " ORDER BY e.id DESC LIMIT 100"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSONError(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type EventSummary struct {
		models.WebhookEvent
		EndpointSlug string `json:"endpoint_slug"`
		EndpointName string `json:"endpoint_name"`
		TargetURL    string `json:"target_url"`
		LatestStatus int    `json:"latest_http_status"`
		LatencyMs    int64  `json:"latest_latency_ms"`
	}

	list := make([]EventSummary, 0)
	for rows.Next() {
		var s EventSummary
		err := rows.Scan(
			&s.ID, &s.EndpointID, &s.RequestID, &s.Method, &s.Path, &s.Headers, &s.QueryParams,
			&s.Body, &s.ContentType, &s.SourceIP, &s.ByteSize, &s.Status, &s.CreatedAt,
			&s.EndpointSlug, &s.EndpointName, &s.TargetURL,
		)
		if err != nil {
			middleware.JSONError(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get latest attempt status
		_ = database.DB.QueryRow(
			"SELECT http_status, latency_ms FROM delivery_attempts WHERE event_id = ? ORDER BY id DESC LIMIT 1",
			s.ID,
		).Scan(&s.LatestStatus, &s.LatencyMs)

		list = append(list, s)
	}

	middleware.JSONSuccess(w, http.StatusOK, list, "webhook events stream retrieved")
}

// GetEvent returns deep detail of single event with all delivery attempts
func GetEvent(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.JSONError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	var d models.EventDetail
	row := database.DB.QueryRow(
		`SELECT e.id, e.endpoint_id, e.request_id, e.method, e.path, e.headers, e.query_params,
		        e.body, e.content_type, e.source_ip, e.byte_size, e.status, e.created_at,
		        ep.slug, ep.name, ep.target_url
		 FROM webhook_events e
		 JOIN endpoints ep ON e.endpoint_id = ep.id
		 WHERE e.id = ?`, id,
	)
	err = row.Scan(
		&d.ID, &d.EndpointID, &d.RequestID, &d.Method, &d.Path, &d.Headers, &d.QueryParams,
		&d.Body, &d.ContentType, &d.SourceIP, &d.ByteSize, &d.Status, &d.CreatedAt,
		&d.EndpointSlug, &d.EndpointName, &d.TargetURL,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.JSONError(w, "event not found", http.StatusNotFound)
			return
		}
		middleware.JSONError(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch attempts
	rows, err := database.DB.Query(
		`SELECT id, event_id, target_url, attempt_num, http_status, response_headers, response_body, latency_ms, error_message, dispatched_at 
		 FROM delivery_attempts WHERE event_id = ? ORDER BY attempt_num DESC`, id,
	)
	if err == nil {
		defer rows.Close()
		d.Attempts = make([]models.DeliveryAttempt, 0)
		for rows.Next() {
			var att models.DeliveryAttempt
			if err := rows.Scan(&att.ID, &att.EventID, &att.TargetURL, &att.AttemptNum,
				&att.HTTPStatus, &att.ResponseHeaders, &att.ResponseBody,
				&att.LatencyMs, &att.ErrorMessage, &att.DispatchedAt); err == nil {
				d.Attempts = append(d.Attempts, att)
			}
		}
	}

	middleware.JSONSuccess(w, http.StatusOK, d, "event inspection detail loaded")
}

// RetryDispatch executes manual or automated HTTP retry for an event
func RetryDispatch(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.JSONError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	var req models.RetryRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	var ev models.WebhookEvent
	var defaultTarget string
	err = database.DB.QueryRow(
		`SELECT e.id, e.endpoint_id, e.request_id, e.method, e.headers, e.body, ep.target_url 
		 FROM webhook_events e
		 JOIN endpoints ep ON e.endpoint_id = ep.id
		 WHERE e.id = ?`, id,
	).Scan(&ev.ID, &ev.EndpointID, &ev.RequestID, &ev.Method, &ev.Headers, &ev.Body, &defaultTarget)

	if err != nil {
		middleware.JSONError(w, "event not found: "+err.Error(), http.StatusNotFound)
		return
	}

	target := defaultTarget
	if req.OverrideTargetURL != "" {
		target = req.OverrideTargetURL
	}
	if target == "" {
		middleware.JSONError(w, "target URL is empty; provide target_url or override", http.StatusBadRequest)
		return
	}

	var maxAttempt int
	_ = database.DB.QueryRow("SELECT COALESCE(MAX(attempt_num), 0) FROM delivery_attempts WHERE event_id = ?", ev.ID).Scan(&maxAttempt)
	nextAttempt := maxAttempt + 1

	headerMap := make(map[string]string)
	_ = json.Unmarshal([]byte(ev.Headers), &headerMap)

	att := executeDispatch(ev.ID, nextAttempt, target, ev.Method, []byte(ev.Body), headerMap)

	newStatus := "RETRIED"
	if att.HTTPStatus >= 200 && att.HTTPStatus < 300 {
		newStatus = "FORWARDED"
		_, _ = database.DB.Exec("UPDATE endpoints SET total_forwarded = total_forwarded + 1 WHERE id = ?", ev.EndpointID)
	} else {
		newStatus = "FAILED"
	}

	_, _ = database.DB.Exec("UPDATE webhook_events SET status = ? WHERE id = ?", newStatus, ev.ID)

	middleware.JSONSuccess(w, http.StatusOK, att, fmt.Sprintf("retry attempt #%d completed with HTTP %d", nextAttempt, att.HTTPStatus))
}

// ListEndpoints returns channels
func ListEndpoints(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(
		`SELECT id, slug, name, description, target_url, secret_token, is_active, total_received, total_forwarded, created_at 
		 FROM endpoints ORDER BY id ASC`,
	)
	if err != nil {
		middleware.JSONError(w, "failed to query endpoints: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	eps := make([]models.Endpoint, 0)
	for rows.Next() {
		var ep models.Endpoint
		if err := rows.Scan(&ep.ID, &ep.Slug, &ep.Name, &ep.Description, &ep.TargetURL,
			&ep.SecretToken, &ep.IsActive, &ep.TotalReceived, &ep.TotalForwarded, &ep.CreatedAt); err == nil {
			eps = append(eps, ep)
		}
	}

	middleware.JSONSuccess(w, http.StatusOK, eps, "endpoints retrieved")
}

// CreateEndpoint registers new channel
func CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var req models.EndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	req.Slug = strings.TrimSpace(strings.ToLower(req.Slug))
	req.Name = strings.TrimSpace(req.Name)
	if req.Slug == "" || req.Name == "" {
		middleware.JSONError(w, "slug and name required", http.StatusBadRequest)
		return
	}

	res, err := database.DB.Exec(
		`INSERT INTO endpoints (slug, name, description, target_url, secret_token, is_active)
		 VALUES (?, ?, ?, ?, ?, 1)`,
		req.Slug, req.Name, req.Description, req.TargetURL, req.SecretToken,
	)
	if err != nil {
		middleware.JSONError(w, "slug already exists or DB failure: "+err.Error(), http.StatusConflict)
		return
	}

	id, _ := res.LastInsertId()
	middleware.JSONSuccess(w, http.StatusCreated, models.Endpoint{
		ID:          id,
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
		TargetURL:   req.TargetURL,
		SecretToken: req.SecretToken,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}, "webhook ingress endpoint registered")
}

// GetStats returns telemetry metrics
func GetStats(w http.ResponseWriter, r *http.Request) {
	var total, delivered, failed, pending int
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM webhook_events").Scan(&total)
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM webhook_events WHERE status = 'FORWARDED'").Scan(&delivered)
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM webhook_events WHERE status = 'FAILED'").Scan(&failed)
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM webhook_events WHERE status = 'RECEIVED'").Scan(&pending)

	var activeEndpoints int
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM endpoints WHERE is_active = 1").Scan(&activeEndpoints)

	stats := models.TelemetryStats{
		TotalEvents:     total,
		DeliveredCount:  delivered,
		FailedCount:     failed,
		PendingCount:    pending,
		ActiveEndpoints: activeEndpoints,
		UptimeSeconds:   int64(time.Since(database.BootTime).Seconds()),
	}

	middleware.JSONSuccess(w, http.StatusOK, stats, "telemetry telemetry stats")
}

// Mock Target: Always returns 200 OK
func MockSuccess(w http.ResponseWriter, r *http.Request) {
	time.Sleep(20 * time.Millisecond) // realistic processing lag
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ACK","committed":true,"receiver":"vault-mock-worker","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

// Mock Target: Always returns 500 Internal Server Error (for demonstration of retry)
func MockFailing(w http.ResponseWriter, r *http.Request) {
	time.Sleep(35 * time.Millisecond)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(`{"error":"Internal Downstream Failure: database lock contention on worker cluster","code":"ERR_TARGET_DEADLOCK"}`))
}

// SimulateTestEvent generates an incoming webhook for testing
func SimulateTestEvent(w http.ResponseWriter, r *http.Request) {
	var req models.TestDispatchReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	slug := "stripe-checkout"
	targetSuffix := "target-success"
	if req.Simulate == "SERVER_ERROR_500" {
		slug = "github-ci-dispatch"
		targetSuffix = "target-failing"
	}

	testPayload := fmt.Sprintf(`{
  "event": "%s",
  "simulation_mode": "%s",
  "generated_at": "%s",
  "payload": {
    "transaction_id": "tx_%d",
    "amount_idr": 1500000,
    "status": "pending_verification"
  }
}`, req.EventName, req.Simulate, time.Now().Format(time.RFC3339), time.Now().Unix())

	target := fmt.Sprintf("http://127.0.0.1:8086/api/mock/%s", targetSuffix)

	// Ingest directly
	var epID int64
	_ = database.DB.QueryRow("SELECT id FROM endpoints WHERE slug = ?", slug).Scan(&epID)
	if epID == 0 {
		epID = 1
	}

	reqID := fmt.Sprintf("REQ-SIM-%04d", time.Now().UnixNano()%10000)
	headers := `{"Content-Type":"application/json","User-Agent":"Vault-Simulation-Engine/1.0"}`

	res, err := database.DB.Exec(
		`INSERT INTO webhook_events (endpoint_id, request_id, method, path, headers, query_params, body, content_type, source_ip, byte_size, status)
		 VALUES (?, ?, 'POST', ?, ?, '{}', ?, 'application/json', '127.0.0.1', ?, 'RECEIVED')`,
		epID, reqID, "/api/v1/ingest/"+slug, headers, testPayload, len(testPayload),
	)
	if err != nil {
		middleware.JSONError(w, "simulation failure: "+err.Error(), http.StatusInternalServerError)
		return
	}

	evID, _ := res.LastInsertId()
	_, _ = database.DB.Exec("UPDATE endpoints SET total_received = total_received + 1 WHERE id = ?", epID)

	headerMap := map[string]string{"Content-Type": "application/json"}
	att := executeDispatch(evID, 1, target, "POST", []byte(testPayload), headerMap)

	status := "RECEIVED"
	if att.HTTPStatus >= 200 && att.HTTPStatus < 300 {
		status = "FORWARDED"
		_, _ = database.DB.Exec("UPDATE endpoints SET total_forwarded = total_forwarded + 1 WHERE id = ?", epID)
	} else {
		status = "FAILED"
	}
	_, _ = database.DB.Exec("UPDATE webhook_events SET status = ? WHERE id = ?", status, evID)

	middleware.JSONSuccess(w, http.StatusCreated, map[string]any{
		"event_id":   evID,
		"request_id": reqID,
		"status":     status,
		"attempt":    att,
	}, "test webhook event simulated and dispatched")
}

// Helper: real HTTP dispatch execution
func executeDispatch(eventID int64, attemptNum int, targetURL, method string, payload []byte, headers map[string]string) models.DeliveryAttempt {
	start := time.Now()

	req, err := http.NewRequest(method, targetURL, bytes.NewBuffer(payload))
	if err != nil {
		att := models.DeliveryAttempt{
			EventID:      eventID,
			TargetURL:    targetURL,
			AttemptNum:   attemptNum,
			HTTPStatus:   0,
			ErrorMessage: "invalid request formation: " + err.Error(),
			LatencyMs:    0,
			DispatchedAt: time.Now(),
		}
		saveAttempt(&att)
		return att
	}

	// Copy headers
	for k, v := range headers {
		if !strings.EqualFold(k, "Host") && !strings.EqualFold(k, "Content-Length") {
			req.Header.Set(k, v)
		}
	}
	req.Header.Set("X-Vault-Delivery-Attempt", strconv.Itoa(attemptNum))
	req.Header.Set("X-Vault-Dispatched-At", time.Now().Format(time.RFC3339))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		att := models.DeliveryAttempt{
			EventID:      eventID,
			TargetURL:    targetURL,
			AttemptNum:   attemptNum,
			HTTPStatus:   0,
			ErrorMessage: "network dispatch error: " + err.Error(),
			LatencyMs:    latency,
			DispatchedAt: time.Now(),
		}
		saveAttempt(&att)
		return att
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	respHeaderMap := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaderMap[k] = v[0]
		}
	}
	respHeaderJSON, _ := json.Marshal(respHeaderMap)

	errMsg := ""
	if resp.StatusCode >= 400 {
		errMsg = fmt.Sprintf("HTTP %d error returned from target", resp.StatusCode)
	}

	att := models.DeliveryAttempt{
		EventID:         eventID,
		TargetURL:       targetURL,
		AttemptNum:      attemptNum,
		HTTPStatus:      resp.StatusCode,
		ResponseHeaders: string(respHeaderJSON),
		ResponseBody:    string(respBody),
		LatencyMs:       latency,
		ErrorMessage:    errMsg,
		DispatchedAt:    time.Now(),
	}
	saveAttempt(&att)
	return att
}

func saveAttempt(att *models.DeliveryAttempt) {
	res, err := database.DB.Exec(
		`INSERT INTO delivery_attempts (event_id, target_url, attempt_num, http_status, response_headers, response_body, latency_ms, error_message, dispatched_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		att.EventID, att.TargetURL, att.AttemptNum, att.HTTPStatus, att.ResponseHeaders, att.ResponseBody, att.LatencyMs, att.ErrorMessage, att.DispatchedAt,
	)
	if err == nil {
		att.ID, _ = res.LastInsertId()
	}
}

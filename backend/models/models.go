package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Endpoint struct {
	ID             int64     `json:"id"`
	Slug           string    `json:"slug"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	TargetURL      string    `json:"target_url"`
	SecretToken    string    `json:"secret_token"`
	IsActive       bool      `json:"is_active"`
	TotalReceived  int       `json:"total_received"`
	TotalForwarded int       `json:"total_forwarded"`
	CreatedAt      time.Time `json:"created_at"`
}

type WebhookEvent struct {
	ID          int64     `json:"id"`
	EndpointID  int64     `json:"endpoint_id"`
	RequestID   string    `json:"request_id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Headers     string    `json:"headers"`      // JSON map
	QueryParams string    `json:"query_params"` // JSON map
	Body        string    `json:"body"`
	ContentType string    `json:"content_type"`
	SourceIP    string    `json:"source_ip"`
	ByteSize    int       `json:"byte_size"`
	Status      string    `json:"status"` // RECEIVED, FORWARDED, FAILED, RETRIED
	CreatedAt   time.Time `json:"created_at"`
}

type DeliveryAttempt struct {
	ID              int64     `json:"id"`
	EventID         int64     `json:"event_id"`
	TargetURL       string    `json:"target_url"`
	AttemptNum      int       `json:"attempt_num"`
	HTTPStatus      int       `json:"http_status"`
	ResponseHeaders string    `json:"response_headers"`
	ResponseBody    string    `json:"response_body"`
	LatencyMs       int64     `json:"latency_ms"`
	ErrorMessage    string    `json:"error_message"`
	DispatchedAt    time.Time `json:"dispatched_at"`
}

type EventDetail struct {
	WebhookEvent
	EndpointSlug string            `json:"endpoint_slug"`
	EndpointName string            `json:"endpoint_name"`
	TargetURL    string            `json:"target_url"`
	Attempts     []DeliveryAttempt `json:"attempts"`
}

type EndpointRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TargetURL   string `json:"target_url"`
	SecretToken string `json:"secret_token"`
}

type RetryRequest struct {
	OverrideTargetURL string `json:"override_target_url,omitempty"`
}

type TestDispatchReq struct {
	EventName string `json:"event_name"`
	Simulate  string `json:"simulate"` // "SUCCESS_200", "SERVER_ERROR_500", "CLIENT_ERROR_404"
}

type AuthRequest struct {
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type TelemetryStats struct {
	TotalEvents     int   `json:"total_events"`
	DeliveredCount  int   `json:"delivered_count"`
	FailedCount     int   `json:"failed_count"`
	PendingCount    int   `json:"pending_count"`
	ActiveEndpoints int   `json:"active_endpoints"`
	UptimeSeconds   int64 `json:"uptime_seconds"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

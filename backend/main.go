package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"webhook-vault/backend/database"
	"webhook-vault/backend/handlers"
	"webhook-vault/backend/middleware"
)

func main() {
	log.Println("[INIT] Booting Webhook Vault Gateway...")

	_, err := database.InitDB()
	if err != nil {
		log.Fatalf("[FATAL] Database initialization failure: %v", err)
	}

	mux := http.NewServeMux()

	// Health Check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "SYSTEM_OPERATIONAL",
			"subsystem": "webhook_vault_engine",
			"version":   "1.0.0-wireframe",
			"protocol":  "RFC-WEBHOOK-INGRESS",
		})
	})

	// Auth Endpoints
	mux.HandleFunc("POST /api/auth/register", handlers.Register)
	mux.HandleFunc("POST /api/auth/login", handlers.Login)
	mux.HandleFunc("GET /api/auth/me", middleware.RequireAuth(handlers.Me))

	// Telemetry & Metrics
	mux.HandleFunc("GET /api/stats", handlers.GetStats)

	// Ingress Endpoints Management
	mux.HandleFunc("GET /api/endpoints", handlers.ListEndpoints)
	mux.HandleFunc("POST /api/endpoints", middleware.RequireAuth(handlers.CreateEndpoint))

	// Webhook Events Stream & Deep Inspection
	mux.HandleFunc("GET /api/events", handlers.ListEvents)
	mux.HandleFunc("GET /api/events/{id}", handlers.GetEvent)

	// Re-dispatch & Retry
	mux.HandleFunc("POST /api/events/{id}/retry", middleware.RequireAuth(handlers.RetryDispatch))
	mux.HandleFunc("POST /api/test/simulate", middleware.RequireAuth(handlers.SimulateTestEvent))

	// Public Webhook Ingress (supports POST, PUT, GET)
	mux.HandleFunc("POST /api/v1/ingest/{slug}", handlers.IngestWebhook)
	mux.HandleFunc("PUT /api/v1/ingest/{slug}", handlers.IngestWebhook)
	mux.HandleFunc("GET /api/v1/ingest/{slug}", handlers.IngestWebhook)

	// Mock Targets for Test & Demonstration of Real HTTP Retries
	mux.HandleFunc("POST /api/mock/target-success", handlers.MockSuccess)
	mux.HandleFunc("POST /api/mock/target-failing", handlers.MockFailing)

	port := getEnv("PORT", "8086")
	addr := fmt.Sprintf("0.0.0.0:%s", port)

	handler := middleware.CORS(mux)

	log.Printf("[READY] Webhook Vault API listening on http://127.0.0.1:%s", port)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("[FATAL] ListenAndServe: %v", err)
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

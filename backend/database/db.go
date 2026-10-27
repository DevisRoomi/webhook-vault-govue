package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB
var BootTime = time.Now()

func InitDB() (*sql.DB, error) {
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASS", "root")
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "webhook_vault")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)

	if err := DB.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := migrate(); err != nil {
		return nil, fmt.Errorf("schema migration: %w", err)
	}

	if err := seed(); err != nil {
		log.Printf("[WARN] seed error: %v", err)
	}

	log.Printf("[DB: OK] MySQL connected: %s:%s/%s", dbHost, dbPort, dbName)
	return DB, nil
}

func migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			email VARCHAR(100) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'OPERATOR',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS endpoints (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			slug VARCHAR(64) NOT NULL UNIQUE,
			name VARCHAR(120) NOT NULL,
			description TEXT,
			target_url VARCHAR(500) NOT NULL,
			secret_token VARCHAR(128) NOT NULL DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			total_received INT NOT NULL DEFAULT 0,
			total_forwarded INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_slug (slug)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS webhook_events (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			endpoint_id BIGINT NOT NULL,
			request_id VARCHAR(64) NOT NULL UNIQUE,
			method VARCHAR(10) NOT NULL,
			path VARCHAR(255) NOT NULL,
			headers LONGTEXT,
			query_params TEXT,
			body LONGTEXT,
			content_type VARCHAR(100) NOT NULL DEFAULT 'application/json',
			source_ip VARCHAR(64) NOT NULL DEFAULT '127.0.0.1',
			byte_size INT NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'RECEIVED',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_endpoint (endpoint_id),
			INDEX idx_status (status),
			INDEX idx_created (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS delivery_attempts (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			event_id BIGINT NOT NULL,
			target_url VARCHAR(500) NOT NULL,
			attempt_num INT NOT NULL DEFAULT 1,
			http_status INT NOT NULL DEFAULT 0,
			response_headers LONGTEXT,
			response_body LONGTEXT,
			latency_ms BIGINT NOT NULL DEFAULT 0,
			error_message TEXT,
			dispatched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_event (event_id),
			INDEX idx_status (http_status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func seed() error {
	var userCount int
	if err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		return err
	}
	if userCount == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("vault2026"), bcrypt.DefaultCost)
		_, _ = DB.Exec(
			"INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)",
			"dispatcher", "operator@vault.wire", string(hash), "OPERATOR",
		)
		log.Println("[SEED] Created user: dispatcher (pass: vault2026)")
	}

	var epCount int
	if err := DB.QueryRow("SELECT COUNT(*) FROM endpoints").Scan(&epCount); err != nil {
		return err
	}
	if epCount == 0 {
		_, _ = DB.Exec(
			`INSERT INTO endpoints (slug, name, description, target_url, secret_token, is_active, total_received, total_forwarded) VALUES 
			('stripe-checkout', 'Stripe Charge Ledger', 'Financial ingress channel capturing payment intent events, invoice signatures, and settlement receipts from Stripe edge gateways.', 'http://127.0.0.1:8086/api/mock/target-success', 'whsec_9b2e84c1f93a', 1, 2, 2),
			('github-ci-dispatch', 'GitHub Actions Dispatch', 'Pipeline dispatch ingress triggering automated integration runs, container validation, and lint builds.', 'http://127.0.0.1:8086/api/mock/target-failing', 'gh_wh_auth_6802x', 1, 1, 0),
			('inventory-stream', 'WMS Ingestion Stream', 'Warehouse management telemetry relay for stock replenishment notices and SKU delta mutations.', 'http://127.0.0.1:8086/api/mock/target-success', 'wms_secret_8819', 1, 1, 1)`,
		)

		// Seed real events
		event1 := `{"id":"evt_3MvwE2LkdIwHu7ix","object":"event","type":"payment_intent.succeeded","created":1761552000,"data":{"object":{"id":"pi_3MvwE2LkdIwHu7ix","amount":4500000,"currency":"idr","status":"succeeded","customer":"cus_9941xa","payment_method":"card_visa"}}}`
		event2 := `{"ref":"refs/heads/main","before":"611a2f","after":"9c88b1","repository":{"name":"webhook-vault-govue","owner":"DevisRoomi"},"commits":[{"id":"9c88b1","message":"feat: raw wireframe ingress buffer"}]}`
		event3 := `{"warehouse_id":"WMS-JKT-01","event":"stock.adjusted","sku":"RAM-DDR5-32G","delta":-4,"balance":142,"audit_id":"AUD-88092"}`

		headers1 := `{"Content-Type":"application/json","User-Agent":"Stripe/1.0 (+https://stripe.com/docs/webhooks)","Stripe-Signature":"t=1761552000,v1=5257abfc"}`
		headers2 := `{"Content-Type":"application/json","User-Agent":"GitHub-Hookshot/9c88b1","X-GitHub-Event":"push","X-GitHub-Delivery":"d19f4a-8801"}`
		headers3 := `{"Content-Type":"application/json","User-Agent":"Go-http-client/1.1","X-Warehouse-Token":"wms_secret_8819"}`

		_, _ = DB.Exec(
			`INSERT INTO webhook_events (endpoint_id, request_id, method, path, headers, query_params, body, content_type, source_ip, byte_size, status, created_at) VALUES 
			(1, 'REQ-8819-001', 'POST', '/api/v1/ingest/stripe-checkout', ?, '{}', ?, 'application/json', '54.187.174.169', 248, 'FORWARDED', NOW() - INTERVAL 12 MINUTE),
			(2, 'REQ-8819-002', 'POST', '/api/v1/ingest/github-ci-dispatch', ?, '{"branch":"main"}', ?, 'application/json', '140.82.115.11', 204, 'FAILED', NOW() - INTERVAL 5 MINUTE),
			(3, 'REQ-8819-003', 'POST', '/api/v1/ingest/inventory-stream', ?, '{}', ?, 'application/json', '10.0.4.12', 134, 'FORWARDED', NOW() - INTERVAL 2 MINUTE)`,
			headers1, event1,
			headers2, event2,
			headers3, event3,
		)

		// Seed real delivery attempts including a real 500 error on event #2
		_, _ = DB.Exec(
			`INSERT INTO delivery_attempts (event_id, target_url, attempt_num, http_status, response_headers, response_body, latency_ms, error_message, dispatched_at) VALUES 
			(1, 'http://127.0.0.1:8086/api/mock/target-success', 1, 200, '{"Content-Type":"application/json","X-Powered-By":"VaultGateway"}', '{"status":"ACK","committed":true}', 42, '', NOW() - INTERVAL 12 MINUTE),
			(2, 'http://127.0.0.1:8086/api/mock/target-failing', 1, 500, '{"Content-Type":"application/json"}', '{"error":"Internal Downstream Failure: database deadlock on build worker queue"}', 318, 'HTTP 500 Internal Server Error returned from target endpoint', NOW() - INTERVAL 5 MINUTE),
			(3, 'http://127.0.0.1:8086/api/mock/target-success', 1, 200, '{"Content-Type":"application/json"}', '{"ack":"SYNC_OK","batch_id":9921}', 38, '', NOW() - INTERVAL 2 MINUTE)`,
		)

		log.Println("[SEED] Seeded 3 webhook endpoints and real event log fixtures (including HTTP 500 downstream error).")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

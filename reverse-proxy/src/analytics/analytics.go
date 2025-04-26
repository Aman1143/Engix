package analytics

import (
	"context"
	"log"
	"time"

	"github.com/Aman1143/reverse-proxy/src/db"
)

func LogRequest(method, path string, status int, latency time.Duration, ip string) {
	err := db.ClickHouseConn.Exec(context.Background(), `
		INSERT INTO analytics.api_logs 
		(method, path, status, latency_ms, ip_address) 
		VALUES (?, ?, ?, ?, ?)`,
		method, path, status, latency.Milliseconds(), ip,
	)
	if err != nil {
		log.Printf("Failed to insert analytics log: %v", err)
	}
}



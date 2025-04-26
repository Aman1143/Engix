package db

import (
	"context"
	"log"
	"time"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
)


var ClickHouseConn clickhouse.Conn

func InitClickHouse() {
	var conn clickhouse.Conn
	var err error

	for retries := 1; retries <= 15; retries++ {
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{"clickhouse:9000"},  
			Auth: clickhouse.Auth{
				Database: "default",
				Username: "default",
				Password: "",
			},
		})
		if err == nil {
			pingErr := conn.Ping(context.Background())
			if pingErr == nil {
				log.Println("✅ Connected to ClickHouse.")
				ClickHouseConn = conn
				break
			}
			log.Printf("Ping error: %v", pingErr)
		} else {
			log.Printf("Connection error: %v", err)
		}

		waitTime := time.Duration(retries*2) * time.Second
		log.Printf("ClickHouse not ready (attempt %d), retrying in %v...", retries, waitTime)
		time.Sleep(waitTime)
	}

	if ClickHouseConn == nil {
		log.Fatal(" Failed to connect to ClickHouse after retries.")
	}
}



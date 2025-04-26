package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Aman1143/reverse-proxy/src/db"
	"github.com/Aman1143/reverse-proxy/src/parse"
	"github.com/Aman1143/reverse-proxy/src/server"
)






func main() {
	if os.Getenv("IS_WORKER") == "true" { 
		db.InitClickHouse()
		validatedConfig := parse.ValidateConfig(parse.ParaseYAMLConfig("config.yaml"))
		server.StartWorker(validatedConfig)
		return
	}

	fmt.Println("🔱 Jai Shree Ram 🔱") 
    
	// db connection
	db.InitClickHouse()
	
	validatedConfig := parse.ValidateConfig(parse.ParaseYAMLConfig("config.yaml"))
	port := validatedConfig.Server.Listen
	workers := validatedConfig.Server.Worker
	if workers == 0 {
		workers = runtime.NumCPU()
	}
    
	server.CreateServer(server.CreateServerConfig{
		Port:        port,
		WorkerCount: workers,
		Config:      validatedConfig,
	})
 
}


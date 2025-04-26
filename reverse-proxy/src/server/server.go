package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Aman1143/reverse-proxy/src/analytics"
	"github.com/Aman1143/reverse-proxy/src/cache"
	"github.com/Aman1143/reverse-proxy/src/configschema"
)

type CreateServerConfig struct {
	Port        int
	WorkerCount int
	Config      configschema.RootConfigSchema
}

type workerHandle struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
}

var nextWorkerIndex uint64 = 0

func getNextWorker(workerCount int) int {
	return int(atomic.AddUint64(&nextWorkerIndex, 1) % uint64(workerCount))
}


func CreateServer(cfg CreateServerConfig) {
	port := cfg.Port
	workerCount := cfg.WorkerCount
	config := cfg.Config

	if os.Getenv("IS_WORKER") == "true" {
		StartWorker(config)
		return
	}

	fmt.Println("Master Process is up")

	workers := make([]workerHandle, workerCount)

	for i := 0; i < workerCount; i++ {
		cmd := exec.Command(os.Args[0])
		cmd.Env = append(os.Environ(),
			"IS_WORKER=true",
			"WORKER_ID="+strconv.Itoa(i),
		)

		stdinPipe, err := cmd.StdinPipe()
		if err != nil {
			log.Fatalf("failed to get stdin for worker %d: %v", i, err)
		}
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("failed to get stdout for worker %d: %v", i, err)
		}
		cmd.Stderr = os.Stderr

		err = cmd.Start()
		if err != nil {
			log.Fatalf("failed to start worker %d: %v", i, err)
		}

		workers[i] = workerHandle{
			cmd:    cmd,
			stdin:  stdinPipe,
			stdout: bufio.NewReader(stdoutPipe),
		}
		fmt.Printf("👷 Worker %d started (PID: %d)\n", i, cmd.Process.Pid)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index := getNextWorker(workerCount)
		worker := workers[index]

		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		} 
		
		payload := configschema.WorkerMessageSchema{
			RequestType: r.Method,
			Headers:     r.Header,
			Body:        string(body),
			URL:         r.URL.String(),
			RemoteAddr:  r.RemoteAddr,
		}
		jsonData, _ := json.Marshal(payload)
		_, err = worker.stdin.Write(append(jsonData, '\n'))

		if err != nil {
			http.Error(w, "Failed to send to worker", http.StatusInternalServerError)
			return
		}

		responseLine, err := worker.stdout.ReadBytes('\n')
		if err != nil {
			http.Error(w, "Failed to read from worker", http.StatusInternalServerError)
			return
		}

		var resp configschema.WorkerResponse
		if err := json.Unmarshal(responseLine, &resp); err != nil {
			log.Printf(" Failed to parse worker response: %s\n", responseLine)
			http.Error(w, "Invalid response from worker", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.Status)
		w.Write([]byte(resp.Body))
	})

	fmt.Printf("📑 Reverse proxy server running at http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal("Server error:", err)
	}
}



func StartWorker(config configschema.RootConfigSchema) {
	cache.InitRedis()

	workerID := os.Getenv("WORKER_ID")
	fmt.Fprintf(os.Stderr, "👷 Worker #%s started with config.Listen: %d\n", workerID, config.Server.Listen)

	reader := bufio.NewReader(os.Stdin)
	requestChan := make(chan configschema.WorkerMessageSchema)

	// Start event-loop-like goroutine to handle requests one by one
	go func() {
		for msg := range requestChan {
			go handleRequest(workerID, msg, config)
		}
	}()

	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Worker #%s read error: %v\n", workerID, err)
			continue
		}

		var msg configschema.WorkerMessageSchema
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Fprintf(os.Stderr, "Worker #%s json error: %v\n", workerID, err)
			continue
		}

		requestChan <- msg
	}
}

func handleRequest(workerID string, msg configschema.WorkerMessageSchema, config configschema.RootConfigSchema) {
	start := time.Now()

	cacheKey := "cache:" + msg.URL
	if msg.RequestType == "GET" {
		cachedData, err := cache.RedisClient.Get(cache.Ctx, cacheKey).Result()
		if err == nil {
			// Cache hit
			fmt.Fprintf(os.Stderr, "cachedDat: %s\n", cachedData)

			// Log analytics for cache hit
			ip :=getIPAddress(msg)
			go analytics.LogRequest(msg.RequestType, msg.URL, http.StatusOK, time.Since(start), ip)

			writeWorkerResponse(workerID, http.StatusOK, cachedData, "")
			return
		}
	}

	parsedURL := strings.TrimPrefix(msg.URL, "/")
	parts := strings.Split(parsedURL, "/")

	var requestPath string
	if len(parts) > 0 {
		requestPath = "/" + parts[0]
	} else {
		requestPath = ""
	}

	targetURL := ""
	for _, rule := range config.Server.Rules {
		if rule.Path == requestPath && len(rule.Upstreams) > 0 {
			for _, up := range config.Server.UpStream {
				if up.ID == rule.Upstreams[0] {
					targetURL = up.URL + msg.URL
					break
				}
			}
			break
		}
	}

	if targetURL == "" {
		writeWorkerResponse(workerID, http.StatusNotFound, "No upstream found for path", "")

		ip := getIPAddress(msg)
		go analytics.LogRequest(msg.RequestType, msg.URL, http.StatusNotFound, time.Since(start), ip)

		return
	}
    
	
	req, err := http.NewRequest(msg.RequestType, targetURL, bytes.NewReader([]byte(msg.Body)))
	// fmt.Printf("Received request: %s %s\n", req.Method, req.URL)

	if err != nil {
		writeWorkerResponse(workerID, http.StatusBadRequest, "Failed to create request", err.Error())

		ip := getIPAddress(msg)

		go analytics.LogRequest(msg.RequestType, msg.URL, http.StatusBadRequest, time.Since(start), ip)

		return
	}
	req.Header = msg.Headers
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("User-Agent", "")

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		writeWorkerResponse(workerID, http.StatusBadGateway, "Upstream request failed", err.Error())

		ip :=getIPAddress(msg)
		go analytics.LogRequest(msg.RequestType, msg.URL, http.StatusBadGateway, time.Since(start), ip)

		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeWorkerResponse(workerID, http.StatusInternalServerError, "Failed to read upstream response", err.Error())

		ip := getIPAddress(msg)
		go analytics.LogRequest(msg.RequestType, msg.URL, http.StatusInternalServerError, time.Since(start), ip)

		return
	}

	bodyStr := string(respBody)

	// Cache only GET responses
	if msg.RequestType == "GET" {
		err = cache.RedisClient.Set(cache.Ctx, cacheKey, bodyStr, 30*time.Second).Err()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set Redis cache: %v\n", err)
		}
	}

	fmt.Fprintf(os.Stderr, "response: %s\n", bodyStr)

	ip :=getIPAddress(msg)

	go analytics.LogRequest(msg.RequestType, msg.URL, resp.StatusCode, time.Since(start), ip)

	writeWorkerResponse(workerID, resp.StatusCode, bodyStr, "")
}




func writeWorkerResponse(workerID string, status int, body string, errMsg string) {
	resp := configschema.WorkerResponse{
		WorkerID: workerID,
		Status:   status,
		Body:     body,
		Error:    errMsg,
	} 
	data, _ := json.Marshal(resp)
	fmt.Printf("%s\n", string(data))  
}


// get ip address
func getIPAddress(msg configschema.WorkerMessageSchema) string {
	IPAddress := ""
	if realIP, exists := msg.Headers["X-Real-Ip"]; exists && len(realIP) > 0 {
		IPAddress = realIP[0]
	}
    if IPAddress == "" {
		if forwardedFor, ok := msg.Headers["X-Forwarded-For"]; ok && len(forwardedFor) > 0 {
			IPAddress = forwardedFor[0]
		}
    }
    if IPAddress == "" {
        IPAddress = msg.RemoteAddr
    }
	return IPAddress
}


Engix 🚀

A high-performance reverse proxy server and custom lightweight CRUD server built using Go, Docker, and ClickHouse — designed for efficient request routing, analytics logging, and concurrency-safe storage.

## Features

- **Reverse Proxy Server**:

  - Handles routing HTTP requests based on path-based rules.
  - Forks multiple worker processes for parallel request processing.
  - Upstream communication using structured JSON-based messaging between master and workers.
  - Integrated real-time analytics logging into ClickHouse and Redis.

- **CRUD Server**:

  - Custom-built file-based database engine supporting `GET`, `POST`, and `DELETE` operations.
  - In-memory concurrency management for simultaneous client requests.
  - Lightweight and database-free (no external DB dependency).

- **Containerized**:
  - Full Docker Compose setup for effortless development and deployment.
  - Services included: Reverse Proxy, CRUD Server, ClickHouse, Redis, Debug container (for quick testing).

## Project Structure

/dbase # CRUD server (Go)
/reverse-proxy # Reverse proxy server (Go)
/clickhouse-init # ClickHouse DB initialization scripts
/docker-compose.yml
/.gitignore
/README.md

## Installation & Setup

1. Clone the repository

```bash
git clone https://github.com/your-username/Engix.git
cd Engix

Build & Run with Docker Compose

bashdocker-compose up --build
This will automatically build the CRUD server and Reverse Proxy, pull Redis and ClickHouse images, and bring up the entire application.

Access Services

Reverse Proxy Server: http://localhost:8080
CRUD Server (directly): http://localhost:3000
ClickHouse UI (optional): http://localhost:8123


Testing an API Example

Through reverse proxy:
bashcurl http://localhost:8080/get/dummy-api
Direct access to CRUD server:
bashcurl http://localhost:3000/get/dummy-api
Requirements

Docker (v20+)
Docker Compose (v2+)
Go (if you want to run servers without Docker)

Future Improvements (optional)

Add HTTPS (TLS) support to reverse proxy.
Implement load balancing between multiple CRUD servers.
Add authentication layer for protected CRUD endpoints.

Made with ❤️ and Jai Shree Ram 🙏
```

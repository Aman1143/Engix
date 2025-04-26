FROM golang:1.23.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY config.yaml .
RUN go mod download

COPY . .

RUN go build -o proxy ./src/main.go
 
FROM debian:bullseye-slim
WORKDIR /app

COPY --from=builder /app/proxy .
COPY --from=builder /app/config.yaml .  

EXPOSE 8080

CMD ["./proxy"]

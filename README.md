# Go-Concurrent-Crawler

![CI](https://github.com/skylerblue333/Go-Concurrent-Crawler/workflows/CI/badge.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg?style=flat&logo=go)
![gRPC](https://img.shields.io/badge/gRPC-Ready-244c5a.svg)

A high-throughput, bounded-concurrency web crawling engine using Go channels, WaitGroups, and semaphore patterns to prevent resource exhaustion.

## System Architecture


```mermaid
graph TD
    Client -->|gRPC/HTTP2| LB[Go Load Balancer]
    LB -->|Round Robin| Node1[Service Node 1]
    LB -->|Round Robin| Node2[Service Node 2]
    Node1 -.->|OpenTelemetry| Jaeger[Jaeger Tracing]
    Node2 -.->|OpenTelemetry| Jaeger
    Node1 <-->|Consul| Discovery[Service Registry]
```


## Elite Features
- **Semaphore Pattern**: Bounded worker pool using buffered channels.
- **Context Cancellation**: Timeout and cancellation propagation across all goroutines.
- **Thread-Safe Deduplication**: `sync.Map` for lock-free URL visited tracking.

## Quick Start
```bash
go mod tidy
go test ./...
go run main.go
```

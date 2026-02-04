# MarkGo Performance

## Overview

MarkGo is designed for high performance with minimal resource usage. This document describes the performance characteristics and how to run benchmarks.

## Performance Characteristics

| Metric | Typical Value |
|--------|---------------|
| Startup time | < 1 second |
| Memory usage | ~30MB RSS |
| Binary size | ~27MB |
| Cached response | < 5ms |
| First request (uncached) | < 50ms |
| Concurrent throughput | 1000+ req/s |

## Comparison

| Platform | Type | Memory | Dependencies |
|----------|------|--------|--------------|
| MarkGo   | Dynamic server | ~30MB | Single binary (~27MB) |
| Ghost    | Dynamic server | ~200MB | Node.js + SQLite |
| WordPress| Dynamic server | ~100MB | PHP + MySQL |
| Hugo     | Static generator | Build-time | Go binary |

## Running Benchmarks

### Go Benchmarks

MarkGo includes micro-benchmarks for config loading, template rendering, and email processing:

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific package benchmarks
go test -bench=. -benchmem ./internal/config/
go test -bench=. -benchmem ./internal/services/
```

### CPU and Memory Profiling

```bash
# CPU profile
go test -cpuprofile=cpu.prof -bench=. ./internal/services/
go tool pprof cpu.prof

# Memory profile
go test -memprofile=mem.prof -bench=. ./internal/services/
go tool pprof mem.prof
```

### Load Testing

For HTTP load testing, use [WebStress](https://github.com/vnykmshr/webstress) (graduated from this project):

```bash
# Install webstress
go install github.com/vnykmshr/webstress@latest

# Run against your MarkGo instance
webstress --url http://localhost:3000
```

## Monitoring in Production

MarkGo exposes runtime metrics:

- `GET /health` - Health check with uptime
- `GET /metrics` - Prometheus-compatible metrics

---

**Last Updated:** February 2026

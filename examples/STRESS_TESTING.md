# Stress Testing for MarkGo

## WebStress - Graduated Stress Testing Tool

The stress testing tool that was previously located in `examples/stress-test/` has been graduated into an independent project called **WebStress**.

### Why the Change?

WebStress has evolved beyond being a MarkGo-specific tool. It's now a general-purpose web stress testing tool that can test any web application, not just MarkGo.

### Where to Find WebStress

- **Repository**: https://github.com/vnykmshr/webstress
- **Documentation**: See the WebStress README for full documentation
- **Installation**:
  ```bash
  go install github.com/vnykmshr/webstress/cmd/webstress@latest
  ```

### Testing MarkGo with WebStress

```bash
# Start your MarkGo server
go run cmd/server/main.go

# In another terminal, run WebStress
webstress -url http://localhost:3000 -duration 2m -concurrency 10 -output markgo-stress-test.json
```

### Key Features

- Automatic URL discovery through crawling
- Concurrent load testing with configurable workers
- Smart rate limiting to respect server capacity
- Comprehensive reporting (HTML, JSON, console)
- Performance validation with pass/fail criteria
- Production-readiness scoring

### Example Configuration for MarkGo

Create a `webstress-config.json`:

```json
{
  "base_url": "http://localhost:3000",
  "concurrency": 10,
  "duration": "5m",
  "rate": 10.0,
  "follow_links": true,
  "max_depth": 3,
  "output_file": "markgo-stress-test.json",
  "performance_targets": {
    "requests_per_second": 100,
    "avg_response_time_ms": 30,
    "p95_response_time_ms": 50,
    "p99_response_time_ms": 100,
    "success_rate": 99.0,
    "error_rate": 1.0
  }
}
```

Then run:
```bash
webstress -config webstress-config.json
```

### Migration Timeline

- **Before**: `examples/stress-test/` in MarkGo repo
- **After**: Independent WebStress project
- **Migration Date**: 2025-10-24
- **MarkGo Version**: Removed in v0.4.0

### Benefits of Independent Project

1. **Broader Use**: Can test any web application
2. **Focused Development**: Dedicated roadmap and features
3. **Community**: Build a community around stress testing
4. **Maintenance**: Easier to maintain and version independently

### Questions or Issues?

- WebStress Issues: https://github.com/vnykmshr/webstress/issues
- MarkGo Issues: https://github.com/vnykmshr/markgo/issues

---

**Note**: The old stress-test code remains in git history if you need to reference it.

# MarkGo Examples

This directory contains example tools and utilities for MarkGo.

## Stress Test Tool

The `stress-test` directory contains a load testing tool for MarkGo server instances.

### Building

```bash
cd stress-test
go build -o stress-test .
```

### Usage

```bash
./stress-test -url http://localhost:3000 -duration 1m -concurrency 10
```

### Options

- `-url`: Target MarkGo server URL (default: http://localhost:3000)
- `-duration`: Test duration (default: 1m)
- `-concurrency`: Number of concurrent users (default: 10)
- `-output`: Output file for results (optional)
- `-verbose`: Verbose logging (optional)

## Note

These tools are not part of the main MarkGo distribution and are maintained separately for development and testing purposes.

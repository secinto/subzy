# Next Steps - Quick Start Guide

## Immediate Actions (This Week)

### 1. Pre-commit Hooks Setup (2 hours)

```bash
# Install pre-commit
pip install pre-commit

# Create .pre-commit-config.yaml
cat > .pre-commit-config.yaml << 'EOF'
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: Go Format
        entry: gofmt -w
        language: system
        files: \.go$

      - id: go-vet
        name: Go Vet
        entry: go vet ./...
        language: system
        pass_filenames: false

      - id: golangci-lint
        name: golangci-lint
        entry: golangci-lint run
        language: system
        pass_filenames: false

      - id: go-test-short
        name: Go Test (short)
        entry: go test -short ./...
        language: system
        pass_filenames: false

      - id: go-mod-tidy
        name: Go Mod Tidy
        entry: go mod tidy
        language: system
        pass_filenames: false
EOF

# Install hooks
pre-commit install

# Test hooks
pre-commit run --all-files
```

### 2. Structured Logging with Graylog (1 week)

**Day 1-2: Setup Dependencies**
```bash
go get -u github.com/rs/zerolog
go get -u github.com/rs/zerolog/log
go get -u gopkg.in/Graylog2/go-gelf.v2/gelf
```

**Day 3-4: Implement Logger**
- Create `runner/logger.go` (see IMPLEMENTATION_PLAN.md)
- Add config fields and CLI flags
- Initialize logger in main.go

**Day 5-7: Replace fmt.Println**
- Update all `fmt.Println` → `logger.Info()`
- Update all `fmt.Printf` → `logger.Debug()`
- Add structured fields (subdomain, engine, status)
- Test Graylog integration

**Testing Graylog Locally:**
```bash
# Docker Compose for local Graylog
docker-compose up -d

# Test with subzy
./subzy run \
  --targets list.txt \
  --log-level debug \
  --log-format json \
  --graylog-host localhost:12201 \
  --graylog-app subzy-test
```

### 3. Increase Test Coverage to 70% (1 week)

**Priority Files:**
1. `process.go` - Add concurrency tests
2. `download.go` - Add HTTP download tests
3. `cmd/run.go` - Add command tests
4. Integration tests - End-to-end workflows

**Commands:**
```bash
# Check current coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# Find untested code
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%"
```

### 4. Fingerprint Matching Optimization (1 week)

**Using Aho-Corasick:**
```bash
# Add dependency
go get -u github.com/cloudflare/ahocorasick

# Create runner/matcher.go
# See IMPLEMENTATION_PLAN.md for full implementation

# Benchmark comparison
go test -bench=BenchmarkMatch -benchmem ./runner
```

**Expected Results:**
- Before: ~100 µs per match (44 fingerprints)
- After: ~10-20 µs per match (5-10x faster)

---

## CLI Enhancements

### New Flags to Add

```go
// Logging
--log-level=info          // debug, info, warn, error
--log-format=console      // console, json
--graylog-host=           // host:port
--graylog-app=subzy       // app name
--log-file                // enable file logging
--log-file-path=subzy.log // log file path

// Retry & Resilience
--max-retries=3           // retry attempts
--retry-backoff=1s        // initial backoff
--timeout-total=5m        // max scan duration

// Rate Limiting
--rate-limit=0            // requests/second (0=unlimited)

// DNS
--dns-check               // check DNS before HTTP
--dns-timeout=5s          // DNS timeout
--dns-only                // DNS-only mode

// Output
--output-format=json      // json, csv, yaml, html, md
--stats                   // print statistics
--no-progress             // disable progress bar

// Security
--insecure-ssl            // skip SSL verification (default: verify)
--ca-cert=                // custom CA certificate

// Advanced
--profile-cpu=cpu.pprof   // CPU profiling
--profile-mem=mem.pprof   // Memory profiling
--custom-fingerprints=    // custom fingerprint files
```

---

## Testing Checklist

### Unit Tests
- [ ] `helpers_test.go` - 100% coverage ✅
- [ ] `worker_test.go` - 90% coverage ✅
- [ ] `fingerprints_test.go` - 95% coverage ✅
- [ ] `reader_test.go` - 95% coverage ✅
- [ ] `config_test.go` - 85% coverage ✅
- [ ] `process_test.go` - Need to add
- [ ] `download_test.go` - Need to add
- [ ] `logger_test.go` - Need to add
- [ ] `matcher_test.go` - Need to add

### Integration Tests
- [ ] End-to-end scan workflow
- [ ] Concurrent processing
- [ ] Error handling
- [ ] Output generation
- [ ] Context cancellation

### Benchmark Tests
- [ ] Fingerprint matching
- [ ] Concurrent processing
- [ ] Memory usage
- [ ] Throughput

---

## Graylog Dashboard Setup

### 1. Install Graylog (Docker)

```yaml
# docker-compose.yml
version: '3'
services:
  mongodb:
    image: mongo:5.0

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.17.9
    environment:
      - "discovery.type=single-node"
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"

  graylog:
    image: graylog/graylog:5.0
    environment:
      - GRAYLOG_HTTP_EXTERNAL_URI=http://127.0.0.1:9000/
      - GRAYLOG_PASSWORD_SECRET=somepasswordpepper
      - GRAYLOG_ROOT_PASSWORD_SHA2=8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918
    ports:
      - "9000:9000"     # Web interface
      - "12201:12201/udp" # GELF UDP
    depends_on:
      - mongodb
      - elasticsearch
```

```bash
docker-compose up -d
```

### 2. Configure Input

1. Open http://localhost:9000
2. Login: admin / admin
3. System → Inputs → Select "GELF UDP" → Launch new input
4. Port: 12201
5. Save

### 3. Create Dashboards

**Scan Overview:**
- Total scans
- Vulnerable subdomains
- Scan duration trends
- Error rate

**Vulnerability Tracking:**
- Vulnerabilities by engine
- New vulnerabilities over time
- Top vulnerable domains

**Performance Metrics:**
- Requests per second
- Average response time
- Error breakdown

---

## Performance Targets

| Metric | Current | Target | Method |
|--------|---------|--------|--------|
| Test Coverage | 35.7% | 70%+ | Add unit & integration tests |
| Fingerprint Match | ~100 µs | ~20 µs | Aho-Corasick algorithm |
| Throughput | ~10-100/s | 500+/s | Optimizations + concurrency |
| Memory Usage | Unbounded | <100MB | Response limits + pooling |
| Error Recovery | 0% | 80%+ | Retry logic |

---

## Development Workflow

### Before Committing
```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run tests
go test -race -cover ./...

# Run benchmarks
go test -bench=. -benchmem ./runner

# Build
go build -o subzy main.go
```

### During Development
```bash
# Watch tests
go install github.com/cespare/reflex@latest
reflex -r '\.go$' -s -- sh -c 'go test ./...'

# Live reload
go install github.com/cosmtrek/air@latest
air
```

### Code Review Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] No breaking changes (or documented)
- [ ] Linter passes
- [ ] Coverage maintained/improved
- [ ] Benchmarks show no regression

---

## Resources

### Documentation
- [Zerolog Guide](https://github.com/rs/zerolog)
- [Graylog Docs](https://docs.graylog.org/)
- [Aho-Corasick Algorithm](https://en.wikipedia.org/wiki/Aho%E2%80%93Corasick_algorithm)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)

### Tools
- [golangci-lint](https://golangci-lint.run/)
- [pre-commit](https://pre-commit.com/)
- [gopls](https://pkg.go.dev/golang.org/x/tools/gopls)
- [delve](https://github.com/go-delve/delve) (debugger)

---

## Questions?

See detailed implementation in `IMPLEMENTATION_PLAN.md`

For phase-by-phase breakdown, refer to sections:
- Phase 2: Testing Foundation
- Phase 3: Performance Optimizations
- Phase 4: Quality & Security (Graylog integration)
- Phase 5: Feature Enhancements
- Phase 6: Advanced Features

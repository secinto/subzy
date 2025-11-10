# Subzy Implementation Plan - Phases 2-6

**Status**: In Progress
**Current Coverage**: 35.7%
**Target Coverage**: 70%+
**Created**: 2025-11-10

---

## Phase 2: Testing Foundation (Week 2-3) ⏳

### Status: 60% Complete

#### ✅ Completed
- [x] Add unit tests for core functions (35.7% coverage)
- [x] Set up GitHub Actions CI/CD
- [x] Add golangci-lint configuration

#### 🔄 In Progress
- [ ] Increase test coverage to 70%+ (Current: 35.7%)
  - [ ] Add tests for `process.go` (complex concurrency logic)
  - [ ] Add tests for `download.go` (HTTP download & file operations)
  - [ ] Add integration tests for end-to-end flows
  - [ ] Add tests for cmd package (command handlers)

#### 📋 TODO
- [ ] **Add pre-commit hooks** (Priority: High)
  - [ ] Install pre-commit framework
  - [ ] Create `.pre-commit-config.yaml`
  - [ ] Configure hooks:
    - `go fmt` - Format code
    - `go vet` - Static analysis
    - `golangci-lint` - Comprehensive linting
    - `go test -short` - Run fast tests
    - `go mod tidy` - Clean dependencies
  - [ ] Document pre-commit setup in README

#### Integration Tests TODO
- [ ] Create `runner/integration_test.go`
  - [ ] Test full scan workflow with mock HTTP server
  - [ ] Test concurrent processing with multiple subdomains
  - [ ] Test JSON output generation
  - [ ] Test error handling in full pipeline
  - [ ] Test rate limiting behavior
  - [ ] Test timeout handling

#### Estimated Time: 1 week

---

## Phase 3: Performance Optimizations (Week 4-5) 🚀

### Status: 40% Complete

#### ✅ Completed
- [x] Add response body size limits (1MB)
- [x] Improve HTTP connection pooling
- [x] Remove dead code
- [x] Basic benchmarks created

#### 📋 TODO

### 3.1 Optimize Fingerprint Matching (Priority: Critical)

**Current**: O(n*m) - Linear scan through all fingerprints for each response
**Target**: O(n+m) - Single pass with efficient matching

#### Option A: Aho-Corasick Algorithm (Recommended)
```go
// File: runner/matcher.go
package runner

import "github.com/cloudflare/ahocorasick"

type FingerprintMatcher struct {
    matcher      *ahocorasick.Matcher
    fingerprints map[int]Fingerprint // Map pattern ID to fingerprint
}

func NewFingerprintMatcher(fingerprints []Fingerprint) *FingerprintMatcher {
    // Build Aho-Corasick automaton from all fingerprint strings
    // Single pass through response body to find all matches
}
```

**Tasks**:
- [ ] Add Aho-Corasick library dependency
- [ ] Create `runner/matcher.go` with new matching logic
- [ ] Implement `NewFingerprintMatcher()` to build trie
- [ ] Implement `Match(body string)` for O(n+m) matching
- [ ] Update `Config.matchResponse()` to use new matcher
- [ ] Add benchmarks comparing old vs new approach
- [ ] Verify no performance regression for small inputs
- [ ] Expected performance: 5-10x faster for typical workloads

#### Option B: Compiled Regex Patterns (Alternative)
```go
type CompiledFingerprint struct {
    Regex          *regexp.Regexp
    FalsePositive  []*regexp.Regexp
    Original       Fingerprint
}
```

**Tasks**:
- [ ] Pre-compile all fingerprint patterns at startup
- [ ] Pre-compile all false positive patterns
- [ ] Use `regexp.FindString()` for matching
- [ ] Benchmark against Aho-Corasick

### 3.2 Advanced Benchmarking & Profiling

**Tasks**:
- [ ] Add CPU profiling support
  - [ ] Create `--profile-cpu` flag
  - [ ] Generate `cpu.pprof` files
  - [ ] Document how to analyze with `go tool pprof`

- [ ] Add memory profiling
  - [ ] Create `--profile-mem` flag
  - [ ] Generate `mem.pprof` files
  - [ ] Identify memory bottlenecks

- [ ] Add comprehensive benchmarks
  - [ ] `BenchmarkFullScanWorkflow` - End-to-end scan
  - [ ] `BenchmarkConcurrentProcessing` - Worker pool performance
  - [ ] `BenchmarkFingerprint100Targets` - Realistic workload
  - [ ] `BenchmarkFingerprint1000Targets` - Stress test

- [ ] Performance regression tests in CI
  - [ ] Add `benchstat` comparison in GitHub Actions
  - [ ] Fail CI if performance degrades >20%

#### Estimated Time: 2 weeks

---

## Phase 4: Quality & Security (Week 6-7) 🔒

### Status: 30% Complete

#### ✅ Completed
- [x] Add input validation (concurrency, timeout, file checks)
- [x] Add configurable User-Agent

#### 📋 TODO

### 4.1 Structured Logging with Graylog Integration (Priority: Critical)

**Architecture**:
```
Application → Zerolog → GELF Format → Graylog Server
                ↓
            Console/File (development)
```

**Implementation**:

#### Step 1: Add Zerolog Dependency
```bash
go get -u github.com/rs/zerolog
go get -u github.com/rs/zerolog/log
go get -u gopkg.in/Graylog2/go-gelf.v2/gelf
```

#### Step 2: Create Logger Package
```go
// File: runner/logger.go
package runner

import (
    "io"
    "os"
    "gopkg.in/Graylog2/go-gelf.v2/gelf"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

type LogConfig struct {
    Level       string // debug, info, warn, error
    Format      string // json, console
    GraylogHost string // e.g., "graylog.example.com:12201"
    GraylogApp  string // Application name
    EnableFile  bool   // Log to file
    FilePath    string // Log file path
}

func InitLogger(cfg LogConfig) (zerolog.Logger, error) {
    var writers []io.Writer

    // Console output (development)
    if cfg.Format == "console" {
        writers = append(writers, zerolog.ConsoleWriter{
            Out: os.Stdout,
            TimeFormat: "15:04:05",
        })
    }

    // Graylog GELF output (production)
    if cfg.GraylogHost != "" {
        gelfWriter, err := gelf.NewUDPWriter(cfg.GraylogHost)
        if err != nil {
            return zerolog.Logger{}, err
        }
        gelfWriter.Facility = cfg.GraylogApp
        writers = append(writers, gelfWriter)
    }

    // File output
    if cfg.EnableFile {
        file, err := os.OpenFile(cfg.FilePath,
            os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
            return zerolog.Logger{}, err
        }
        writers = append(writers, file)
    }

    multi := zerolog.MultiLevelWriter(writers...)
    logger := zerolog.New(multi).With().
        Timestamp().
        Str("app", cfg.GraylogApp).
        Logger()

    // Set log level
    level, err := zerolog.ParseLevel(cfg.Level)
    if err != nil {
        level = zerolog.InfoLevel
    }
    logger = logger.Level(level)

    return logger, nil
}
```

#### Step 3: Add Config Fields
```go
// File: runner/config.go
type Config struct {
    // ... existing fields ...

    // Logging
    LogLevel     string
    LogFormat    string
    GraylogHost  string
    GraylogApp   string
    LogToFile    bool
    LogFilePath  string

    logger       zerolog.Logger
}
```

#### Step 4: Update cmd/run.go with Flags
```go
runCmd.Flags().StringVar(&opts.LogLevel, "log-level", "info",
    "Log level: debug, info, warn, error")
runCmd.Flags().StringVar(&opts.LogFormat, "log-format", "console",
    "Log format: json, console")
runCmd.Flags().StringVar(&opts.GraylogHost, "graylog-host", "",
    "Graylog server host:port (e.g., graylog.example.com:12201)")
runCmd.Flags().StringVar(&opts.GraylogApp, "graylog-app", "subzy",
    "Application name for Graylog")
runCmd.Flags().BoolVar(&opts.LogToFile, "log-file", false,
    "Enable logging to file")
runCmd.Flags().StringVar(&opts.LogFilePath, "log-file-path", "subzy.log",
    "Log file path")
```

#### Step 5: Integrate Logging Throughout Application
```go
// Example usage in process.go
logger.Info().
    Int("subdomain_count", len(subdomains)).
    Int("fingerprint_count", len(fingerprints)).
    Int("concurrency", config.Concurrency).
    Msg("Starting subdomain scan")

logger.Debug().
    Str("subdomain", subdomain).
    Msg("Checking subdomain")

logger.Warn().
    Str("subdomain", subdomain).
    Err(err).
    Msg("HTTP request failed")

logger.Error().
    Str("subdomain", subdomain).
    Str("engine", result.entry.Engine).
    Msg("Vulnerable subdomain detected")
```

**Tasks**:
- [ ] Add zerolog and go-gelf dependencies
- [ ] Create `runner/logger.go` with initialization
- [ ] Add logging config fields to `Config` struct
- [ ] Add CLI flags for logging configuration
- [ ] Replace all `fmt.Println` with structured logging
- [ ] Add contextual logging (subdomain, engine, status)
- [ ] Create Graylog dashboard examples
- [ ] Document logging setup in README
- [ ] Add logging tests

### 4.2 Progress Indicators

**Library**: github.com/schollz/progressbar/v3

**Tasks**:
- [ ] Add progressbar dependency
- [ ] Create progress bar for subdomain scanning
- [ ] Show: `[████████░░] 80/100 subdomains | 5 vulnerable | 2.5/s`
- [ ] Add `--no-progress` flag to disable
- [ ] Ensure progress bar works with logging
- [ ] Update when writing to JSON

### 4.3 Retry Logic with Exponential Backoff

**Implementation**:
```go
// File: runner/retry.go
package runner

import (
    "context"
    "math"
    "time"
)

type RetryConfig struct {
    MaxRetries     int
    InitialBackoff time.Duration
    MaxBackoff     time.Duration
    Multiplier     float64
}

func (c *Config) checkSubdomainWithRetry(ctx context.Context, subdomain string) Result {
    var result Result
    backoff := c.RetryConfig.InitialBackoff

    for attempt := 0; attempt <= c.RetryConfig.MaxRetries; attempt++ {
        result = c.checkSubdomain(subdomain)

        if result.resStatus != ResultHTTPError {
            return result // Success or non-retryable error
        }

        if attempt < c.RetryConfig.MaxRetries {
            c.logger.Debug().
                Str("subdomain", subdomain).
                Int("attempt", attempt+1).
                Dur("backoff", backoff).
                Msg("Retrying after error")

            select {
            case <-time.After(backoff):
                backoff = time.Duration(float64(backoff) * c.RetryConfig.Multiplier)
                if backoff > c.RetryConfig.MaxBackoff {
                    backoff = c.RetryConfig.MaxBackoff
                }
            case <-ctx.Done():
                return result // Context cancelled
            }
        }
    }

    return result
}
```

**Tasks**:
- [ ] Create `runner/retry.go`
- [ ] Add `RetryConfig` to main `Config`
- [ ] Add CLI flags: `--max-retries`, `--retry-backoff`
- [ ] Update `checkSubdomain` to use retry logic
- [ ] Add retry metrics (attempts, backoff times)
- [ ] Test retry behavior
- [ ] Default: 3 retries, 1s initial, 10s max, 2x multiplier

### 4.4 Context Support for Cancellation

**Tasks**:
- [ ] Update `Process()` signature: `func Process(ctx context.Context, config *Config) error`
- [ ] Propagate context to all workers
- [ ] Listen for `ctx.Done()` in worker loops
- [ ] Handle Ctrl+C gracefully (SIGINT, SIGTERM)
- [ ] Add `--timeout-total` flag for max scan duration
- [ ] Clean up resources on cancellation
- [ ] Save partial results before exit

**Example**:
```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle signals
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-sigCh
        cancel()
    }()

    if err := runner.Process(ctx, &opts); err != nil {
        // ...
    }
}
```

### 4.5 Secure TLS Defaults

**Current Issue**: Defaults to `InsecureSkipVerify: true`

**Tasks**:
- [ ] Change default: `VerifySSL: true`
- [ ] Update flag description: `--insecure-ssl` to explicitly skip verification
- [ ] Add warning when using `--insecure-ssl`
- [ ] Add certificate validation logging
- [ ] Support custom CA certificates
- [ ] Add `--ca-cert` flag for custom root CAs

### 4.6 Rate Limiting

**Library**: golang.org/x/time/rate

**Implementation**:
```go
// File: runner/ratelimit.go
package runner

import (
    "context"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
    }
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    return rl.limiter.Wait(ctx)
}
```

**Tasks**:
- [ ] Add rate limiting library
- [ ] Create `runner/ratelimit.go`
- [ ] Add `--rate-limit` flag (requests/second)
- [ ] Integrate rate limiter in worker loop
- [ ] Add per-domain rate limiting option
- [ ] Add rate limit bypass flag for testing
- [ ] Default: No limit (backwards compatible)

#### Estimated Time: 2 weeks

---

## Phase 5: Feature Enhancements (Week 8-10) ⭐

### Status: 20% Complete

#### ✅ Completed
- [x] Add configurable User-Agent

#### 📋 TODO

### 5.1 DNS Checking Before HTTP

**Purpose**: Reduce false positives by checking DNS records first

**Implementation**:
```go
// File: runner/dns.go
package runner

import (
    "context"
    "net"
    "time"
)

type DNSResult struct {
    HasCNAME    bool
    CNAME       string
    HasA        bool
    ARecords    []string
    IsDangling  bool // CNAME exists but no A records
}

func (c *Config) checkDNS(ctx context.Context, subdomain string) (*DNSResult, error) {
    resolver := &net.Resolver{
        PreferGo: true,
        Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
            d := net.Dialer{Timeout: 5 * time.Second}
            return d.DialContext(ctx, network, address)
        },
    }

    // Check CNAME
    cname, err := resolver.LookupCNAME(ctx, subdomain)

    // Check A records
    ips, err := resolver.LookupHost(ctx, subdomain)

    return &DNSResult{
        HasCNAME:   cname != subdomain,
        CNAME:      cname,
        HasA:       len(ips) > 0,
        ARecords:   ips,
        IsDangling: (cname != subdomain) && (len(ips) == 0),
    }, nil
}
```

**Tasks**:
- [ ] Create `runner/dns.go`
- [ ] Add DNS checking before HTTP requests
- [ ] Add `--dns-check` flag (default: true)
- [ ] Add `--dns-timeout` flag
- [ ] Detect dangling CNAME records
- [ ] Add DNS results to JSON output
- [ ] Cache DNS results to avoid duplicate lookups
- [ ] Add DNS-only mode (skip HTTP)

### 5.2 Multiple Output Formats

**Supported Formats**:
- JSON (existing)
- CSV
- YAML
- HTML (web report)
- Markdown (for documentation)
- Plain text (simple list)

**Implementation**:
```go
// File: runner/output.go
package runner

type OutputFormatter interface {
    Format(results []*subdomainResult) ([]byte, error)
}

type JSONFormatter struct{}
type CSVFormatter struct{}
type YAMLFormatter struct{}
type HTMLFormatter struct{}
type MarkdownFormatter struct{}
```

**Tasks**:
- [ ] Create `runner/output.go` with formatter interface
- [ ] Implement JSON formatter (refactor existing)
- [ ] Implement CSV formatter
- [ ] Implement YAML formatter
- [ ] Implement HTML formatter with styling
- [ ] Implement Markdown formatter
- [ ] Add `--output-format` flag
- [ ] Auto-detect format from file extension
- [ ] Add templates for HTML/Markdown
- [ ] Support multiple output files simultaneously

### 5.3 Statistics and Summary Reporting

**Metrics to Track**:
- Total subdomains scanned
- Vulnerable count
- Not vulnerable count
- HTTP errors count
- Response errors count
- Scan duration
- Average response time
- Requests per second
- Success rate
- Unique engines found

**Implementation**:
```go
// File: runner/stats.go
package runner

import "time"

type ScanStatistics struct {
    TotalScanned     int
    Vulnerable       int
    NotVulnerable    int
    HTTPErrors       int
    ResponseErrors   int
    StartTime        time.Time
    EndTime          time.Time
    Duration         time.Duration
    AvgResponseTime  time.Duration
    RequestsPerSec   float64
    SuccessRate      float64
    EnginesFound     map[string]int
}

func (s *ScanStatistics) Print() {
    fmt.Println("\n=== Scan Summary ===")
    fmt.Printf("Total Scanned:    %d\n", s.TotalScanned)
    fmt.Printf("Vulnerable:       %d (%.1f%%)\n", s.Vulnerable, ...)
    fmt.Printf("Not Vulnerable:   %d\n", s.NotVulnerable)
    fmt.Printf("Errors:           %d\n", s.HTTPErrors+s.ResponseErrors)
    fmt.Printf("Duration:         %s\n", s.Duration)
    fmt.Printf("Throughput:       %.2f req/s\n", s.RequestsPerSec)
    // ...
}
```

**Tasks**:
- [ ] Create `runner/stats.go`
- [ ] Track metrics during scan
- [ ] Print summary at end of scan
- [ ] Add `--stats` flag to enable/disable
- [ ] Add statistics to JSON output
- [ ] Export stats to separate file
- [ ] Create charts/graphs for HTML output
- [ ] Real-time stats with `--live-stats` flag

### 5.4 Plugin System for Custom Fingerprints

**Architecture**:
```
~/.subzy/
├── fingerprints.json          # Official fingerprints
├── custom/
│   ├── company-services.json  # Custom fingerprints
│   ├── internal-apps.json
│   └── legacy-systems.json
└── plugins/
    └── fingerprint-loader.so  # Optional: Go plugins
```

**Implementation**:
```go
// File: runner/plugins.go
package runner

type FingerprintSource interface {
    Load() ([]Fingerprint, error)
    Name() string
}

type LocalFileSource struct {
    Path string
}

type RemoteURLSource struct {
    URL string
}

type PluginManager struct {
    sources []FingerprintSource
}

func (pm *PluginManager) LoadAllFingerprints() ([]Fingerprint, error) {
    var all []Fingerprint
    for _, source := range pm.sources {
        fps, err := source.Load()
        if err != nil {
            log.Warn().Err(err).Str("source", source.Name()).Msg("Failed to load")
            continue
        }
        all = append(all, fps...)
    }
    return all, nil
}
```

**Tasks**:
- [ ] Create `runner/plugins.go`
- [ ] Support loading from multiple JSON files
- [ ] Add `--custom-fingerprints` flag (comma-separated paths)
- [ ] Auto-load from `~/.subzy/custom/` directory
- [ ] Support remote fingerprint URLs
- [ ] Add fingerprint validation
- [ ] Merge duplicate fingerprints
- [ ] Add fingerprint priority/ordering
- [ ] Create fingerprint testing tool
- [ ] Document custom fingerprint format

#### Estimated Time: 3 weeks

---

## Phase 6: Advanced Features (Future) 🚀

### Status: 0% Complete (Planning Only)

### 6.1 API Server Mode

**Purpose**: Run Subzy as a RESTful API service

**Tech Stack**:
- Framework: `gin-gonic/gin` or `gorilla/mux`
- Authentication: JWT tokens
- Rate limiting: Per API key
- Database: PostgreSQL or SQLite for results

**Endpoints**:
```
POST   /api/v1/scans              - Create new scan
GET    /api/v1/scans/:id          - Get scan status
GET    /api/v1/scans/:id/results  - Get scan results
GET    /api/v1/scans              - List all scans
DELETE /api/v1/scans/:id          - Delete scan
GET    /api/v1/health             - Health check
GET    /api/v1/metrics            - Prometheus metrics
```

**Tasks**:
- [ ] Design API specification (OpenAPI/Swagger)
- [ ] Create `server/` package
- [ ] Implement HTTP server with Gin
- [ ] Add JWT authentication
- [ ] Implement scan queue with workers
- [ ] Add database layer for persistence
- [ ] Add WebSocket for real-time updates
- [ ] Add API rate limiting
- [ ] Create Swagger documentation
- [ ] Add Prometheus metrics endpoint
- [ ] Docker containerization
- [ ] Kubernetes deployment manifests

### 6.2 Web Dashboard

**Purpose**: Web UI for managing scans and viewing results

**Tech Stack**:
- Frontend: React + TypeScript
- UI Library: Material-UI or Tailwind CSS
- Charts: Recharts or Chart.js
- State: Redux or Zustand
- Build: Vite

**Features**:
- Dashboard with scan statistics
- Submit new scans
- View scan results in table
- Filter and search results
- Export results (CSV, JSON, PDF)
- User management
- Scan scheduling
- Historical trends
- Vulnerability timeline

**Tasks**:
- [ ] Create `web/` directory
- [ ] Set up React + TypeScript project
- [ ] Design UI mockups
- [ ] Implement dashboard view
- [ ] Implement scan submission form
- [ ] Implement results table with filters
- [ ] Add charts and visualizations
- [ ] Integrate with API backend
- [ ] Add authentication flow
- [ ] Responsive design
- [ ] Dark mode support
- [ ] Build and embed into Go binary

### 6.3 Historical Tracking

**Purpose**: Track vulnerability changes over time

**Database Schema**:
```sql
CREATE TABLE scans (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP,
    target_count INT,
    vulnerable_count INT,
    duration_ms INT,
    status VARCHAR(20)
);

CREATE TABLE scan_results (
    id UUID PRIMARY KEY,
    scan_id UUID REFERENCES scans(id),
    subdomain VARCHAR(255),
    status VARCHAR(50),
    engine VARCHAR(100),
    created_at TIMESTAMP
);

CREATE TABLE vulnerability_history (
    subdomain VARCHAR(255),
    engine VARCHAR(100),
    first_seen TIMESTAMP,
    last_seen TIMESTAMP,
    scan_count INT,
    PRIMARY KEY (subdomain, engine)
);
```

**Features**:
- Store all scan results
- Track when vulnerabilities first appear
- Track when vulnerabilities are fixed
- Generate trend reports
- Alert on new vulnerabilities
- Compare scans over time

**Tasks**:
- [ ] Choose database (PostgreSQL recommended)
- [ ] Create database schema
- [ ] Add database migrations
- [ ] Implement data access layer
- [ ] Store scan results automatically
- [ ] Add `--save-history` flag
- [ ] Query historical data
- [ ] Generate trend reports
- [ ] Add `subzy history` command
- [ ] Export historical data

### 6.4 Notification Integrations

**Supported Channels**:
- Slack
- Discord
- Microsoft Teams
- Email (SMTP)
- PagerDuty
- Webhooks (generic)
- Telegram

**Implementation**:
```go
// File: runner/notifications.go
package runner

type Notifier interface {
    Notify(result *NotificationPayload) error
}

type NotificationPayload struct {
    Severity    string
    Subdomain   string
    Engine      string
    Description string
    Timestamp   time.Time
}

type SlackNotifier struct {
    WebhookURL string
}

type DiscordNotifier struct {
    WebhookURL string
}

// ... etc
```

**Tasks**:
- [ ] Create `runner/notifications.go`
- [ ] Implement Slack notifier
- [ ] Implement Discord notifier
- [ ] Implement Email notifier
- [ ] Implement webhook notifier
- [ ] Add notification config file
- [ ] Add `--notify-on` flag (vulnerable, error, all)
- [ ] Template system for messages
- [ ] Batch notifications
- [ ] Test all notification channels

### 6.5 ML-Based Detection

**Purpose**: Use machine learning to improve detection accuracy

**Approach**:
1. **Feature Engineering**:
   - Response body length
   - Response headers
   - Status code
   - Response time
   - HTML structure
   - JavaScript presence
   - CSS patterns
   - DNS records

2. **Model Training**:
   - Collect labeled dataset (vulnerable vs not)
   - Train classifier (Random Forest, XGBoost, or Neural Network)
   - Validate on test set
   - Deploy model

3. **Inference**:
   - Extract features from response
   - Run model prediction
   - Combine with fingerprint matching
   - Confidence scoring

**Tasks**:
- [ ] Research existing takeover datasets
- [ ] Collect training data
- [ ] Feature extraction implementation
- [ ] Model training pipeline (Python + scikit-learn)
- [ ] Model export (ONNX or TensorFlow Lite)
- [ ] Go ML inference library integration
- [ ] Confidence scoring system
- [ ] False positive learning
- [ ] Model versioning and updates
- [ ] A/B testing framework

#### Estimated Time: 12+ weeks (long-term roadmap)

---

## Implementation Priority Matrix

| Feature | Priority | Complexity | Impact | Est. Time |
|---------|----------|------------|--------|-----------|
| Pre-commit hooks | High | Low | Medium | 2 days |
| Test coverage to 70% | High | Medium | High | 1 week |
| Structured logging + Graylog | Critical | Medium | High | 1 week |
| Fingerprint optimization | Critical | High | Very High | 1 week |
| Context cancellation | High | Medium | Medium | 3 days |
| Progress indicators | Medium | Low | Medium | 2 days |
| Retry logic | High | Medium | High | 3 days |
| Rate limiting | High | Medium | High | 3 days |
| DNS checking | Medium | Medium | High | 1 week |
| Statistics reporting | Medium | Low | Medium | 3 days |
| Multiple output formats | Medium | Medium | Medium | 1 week |
| Secure TLS defaults | High | Low | High | 2 days |
| Plugin system | Low | High | Medium | 2 weeks |
| API server mode | Low | Very High | Medium | 4+ weeks |
| Web dashboard | Low | Very High | Medium | 6+ weeks |
| Historical tracking | Low | High | Medium | 3 weeks |
| Notifications | Low | Medium | Medium | 2 weeks |
| ML detection | Low | Very High | High | 12+ weeks |

---

## Quick Start Roadmap (Next 4 Weeks)

### Week 1
- [ ] Pre-commit hooks setup
- [ ] Increase test coverage to 50%
- [ ] Start structured logging implementation

### Week 2
- [ ] Complete structured logging + Graylog
- [ ] Implement context cancellation
- [ ] Add progress indicators
- [ ] Increase test coverage to 70%

### Week 3
- [ ] Fingerprint matching optimization (Aho-Corasick)
- [ ] Retry logic with exponential backoff
- [ ] Rate limiting

### Week 4
- [ ] DNS checking
- [ ] Secure TLS defaults
- [ ] Statistics reporting
- [ ] Performance benchmarking

---

## Success Metrics

### Phase 2-3 (Testing & Performance)
- ✅ Test coverage ≥ 70%
- ✅ All CI checks passing
- ✅ Pre-commit hooks functional
- ✅ 5-10x faster fingerprint matching
- ✅ Memory usage stable under load

### Phase 4 (Quality & Security)
- ✅ Structured logging operational
- ✅ Graylog integration working
- ✅ Retry logic reduces transient failures by 80%+
- ✅ Graceful cancellation works
- ✅ TLS verification enabled by default

### Phase 5 (Features)
- ✅ DNS checking reduces false positives
- ✅ 5+ output formats supported
- ✅ Statistics provide actionable insights
- ✅ Plugin system allows custom fingerprints

### Phase 6 (Advanced)
- ✅ API server handles 100+ concurrent scans
- ✅ Web dashboard fully functional
- ✅ Historical tracking operational
- ✅ Notifications working for all channels
- ✅ ML model improves accuracy by 10%+

---

## Dependencies to Add

```bash
# Phase 4: Logging
go get -u github.com/rs/zerolog
go get -u gopkg.in/Graylog2/go-gelf.v2/gelf

# Phase 4: Progress
go get -u github.com/schollz/progressbar/v3

# Phase 4: Rate Limiting
go get -u golang.org/x/time/rate

# Phase 3: Fingerprint Matching
go get -u github.com/cloudflare/ahocorasick
# OR
go get -u github.com/anknown/ahocorasick

# Phase 5: Output Formats
go get -u gopkg.in/yaml.v3
go get -u github.com/olekukonko/tablewriter

# Phase 6: API Server
go get -u github.com/gin-gonic/gin
go get -u github.com/golang-jwt/jwt/v5
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres

# Phase 6: ML
go get -u github.com/owulveryck/onnx-go
```

---

## Documentation Updates Needed

- [ ] README: Add Graylog setup instructions
- [ ] README: Add pre-commit hooks setup
- [ ] README: Add all new CLI flags
- [ ] CONTRIBUTING.md: Add testing guidelines
- [ ] CONTRIBUTING.md: Add logging guidelines
- [ ] ARCHITECTURE.md: New document explaining design
- [ ] API.md: API documentation (Phase 6)
- [ ] DEPLOYMENT.md: Docker/Kubernetes guide (Phase 6)

---

## Notes

- All phases build incrementally on previous work
- Phases 2-4 should be completed before Phase 5
- Phase 6 is long-term and can be done in parallel once core is stable
- Backward compatibility must be maintained
- All new features should have tests
- All breaking changes require major version bump

---

**Last Updated**: 2025-11-10
**Next Review**: Every 2 weeks during active development

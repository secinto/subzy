# Graylog Integration Guide

This guide explains how to set up and use Graylog logging with Subzy.

## Table of Contents

- [Quick Start](#quick-start)
- [Local Development Setup](#local-development-setup)
- [Production Setup](#production-setup)
- [Usage Examples](#usage-examples)
- [Dashboard Configuration](#dashboard-configuration)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

### 1. Start Local Graylog (Docker)

```bash
# Start Graylog, MongoDB, and Elasticsearch
docker-compose up -d

# Wait for services to start (about 60 seconds)
docker-compose logs -f graylog

# When you see "Graylog server up and running", press Ctrl+C
```

### 2. Configure Graylog Input

1. Open http://localhost:9000
2. Login: `admin` / `admin`
3. Go to **System** → **Inputs**
4. Select **GELF UDP** from dropdown
5. Click **Launch new input**
6. Configure:
   - **Title**: Subzy
   - **Port**: 12201 (default)
   - **Bind address**: 0.0.0.0
7. Click **Save**

### 3. Run Subzy with Graylog

```bash
# Run with Graylog logging
./subzy run \
  --target example.com \
  --log-level debug \
  --graylog-host localhost:12201

# View logs in Graylog web UI
```

---

## Local Development Setup

### Prerequisites

- Docker and Docker Compose
- 4GB+ RAM available for containers
- Ports 9000, 12201 available

### Full Setup

```bash
# Clone repository
git clone https://github.com/LukaSikic/subzy.git
cd subzy

# Start Graylog stack
docker-compose up -d

# Check services are running
docker-compose ps

# Expected output:
# subzy-graylog         Up      9000/tcp, 12201/udp
# subzy-mongodb         Up      27017/tcp
# subzy-elasticsearch   Up      9200/tcp

# Wait for Graylog to be ready
docker-compose logs -f graylog
# Look for: "Graylog server up and running"

# Build subzy
make build

# Test logging
./subzy run --target test.example.com --graylog-host localhost:12201 --log-level debug
```

### Accessing Graylog

- **URL**: http://localhost:9000
- **Username**: admin
- **Password**: admin

---

## Production Setup

### On Existing Graylog Server

If you already have a Graylog server:

```bash
# Run subzy with your Graylog server
./subzy run \
  --targets domains.txt \
  --graylog-host graylog.company.com:12201 \
  --graylog-app subzy-prod \
  --log-level info
```

### Security Considerations

1. **Change default password** in Graylog
2. **Update GRAYLOG_PASSWORD_SECRET** in docker-compose.yml
3. **Use TLS** for production GELF connections
4. **Restrict network access** to Graylog ports
5. **Enable authentication** if exposing publicly

### Production docker-compose.yml

```yaml
version: '3'

services:
  graylog:
    image: graylog/graylog:5.0
    environment:
      - GRAYLOG_PASSWORD_SECRET=${GRAYLOG_SECRET}  # From .env file
      - GRAYLOG_ROOT_PASSWORD_SHA2=${GRAYLOG_PASSWORD_HASH}
      - GRAYLOG_HTTP_EXTERNAL_URI=https://graylog.company.com/
      # ... other settings
    volumes:
      - /opt/graylog/data:/usr/share/graylog/data
    ports:
      - "127.0.0.1:9000:9000"  # Only localhost access
      - "12201:12201/udp"      # GELF input
```

---

## Usage Examples

### Basic Logging

```bash
# Console logging (default)
./subzy run --target example.com

# JSON logging to stdout
./subzy run --target example.com --log-format json

# Debug level logging
./subzy run --target example.com --log-level debug
```

### Graylog Integration

```bash
# Send logs to Graylog only
./subzy run \
  --targets domains.txt \
  --graylog-host graylog.internal:12201 \
  --graylog-app subzy-scanner \
  --log-format json

# Graylog + console output
./subzy run \
  --targets domains.txt \
  --graylog-host localhost:12201 \
  --log-format console

# Graylog + file logging
./subzy run \
  --targets domains.txt \
  --graylog-host localhost:12201 \
  --log-file \
  --log-file-path /var/log/subzy/scan.log
```

### Multi-Output Logging

```bash
# Console + Graylog + File
./subzy run \
  --targets domains.txt \
  --log-level info \
  --log-format console \
  --graylog-host localhost:12201 \
  --graylog-app subzy \
  --log-file \
  --log-file-path subzy-$(date +%Y%m%d).log
```

### Production Scanning

```bash
# Production scan with all logging
./subzy run \
  --targets /opt/subzy/domains.txt \
  --output /opt/subzy/results/scan-$(date +%Y%m%d-%H%M).json \
  --graylog-host graylog.internal:12201 \
  --graylog-app subzy-prod \
  --log-level info \
  --log-file \
  --log-file-path /var/log/subzy/scan-$(date +%Y%m%d-%H%M).log \
  --concurrency 50 \
  --timeout 15 \
  --verify_ssl
```

---

## Dashboard Configuration

### Creating Dashboards in Graylog

#### 1. Scan Overview Dashboard

1. Go to **Dashboards** → **Create dashboard**
2. Name: "Subzy Scan Overview"
3. Add widgets:

**Total Scans** (Count)
- **Search**: `app:subzy AND message:"Starting subdomain takeover scan"`
- **Type**: Count
- **Time Range**: Last 24 hours

**Vulnerable Subdomains** (Count)
- **Search**: `app:subzy AND message:"Vulnerable subdomain detected"`
- **Type**: Count
- **Time Range**: Last 24 hours

**Scan Duration** (Stats)
- **Search**: `app:subzy AND message:"Scan completed"`
- **Type**: Statistics
- **Field**: `duration_ms`

**Vulnerabilities by Engine** (Pie Chart)
- **Search**: `app:subzy AND status:vulnerable`
- **Type**: Pie chart
- **Field**: `engine`

#### 2. Real-Time Monitoring Dashboard

**Live Vulnerability Feed** (Message Table)
- **Search**: `app:subzy AND level:error`
- **Type**: Message table
- **Fields**: timestamp, subdomain, engine, documentation
- **Sort**: Timestamp desc

**Error Rate** (Line Chart)
- **Search**: `app:subzy AND level:error`
- **Type**: Line chart
- **Interval**: 5 minutes

**Scan Progress** (Count)
- **Search**: `app:subzy AND message:"Subdomain check completed"`
- **Type**: Count
- **Time Range**: Last 1 hour

### Example Graylog Queries

```
# All vulnerabilities
app:subzy AND status:vulnerable

# Vulnerabilities for specific service
app:subzy AND status:vulnerable AND engine:"AWS S3"

# HTTP errors
app:subzy AND status:"http error"

# Specific subdomain
app:subzy AND subdomain:"test.example.com"

# High-level errors only
app:subzy AND level:error

# Scans from specific instance
app:subzy AND graylog_app:"subzy-prod"

# Time range
app:subzy AND timestamp:[2025-01-01 TO 2025-01-31]
```

### Alerts Configuration

#### Alert: New Vulnerability Detected

1. Go to **Alerts** → **Event Definitions** → **Create Event Definition**
2. Configure:
   - **Title**: New Subdomain Vulnerability
   - **Priority**: High
   - **Condition**: Filter & Aggregation
   - **Search Query**: `app:subzy AND status:vulnerable`
   - **Aggregation**: count()
   - **Threshold**: >= 1
   - **Time Range**: 5 minutes
   - **Notification**: Email/Slack

#### Alert: High Error Rate

1. **Title**: Subzy High Error Rate
2. **Condition**: `app:subzy AND level:error`
3. **Aggregation**: count()
4. **Threshold**: >= 10 in 5 minutes

---

## Structured Log Fields

Subzy sends these structured fields to Graylog:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `app` | string | Application name | `subzy` |
| `level` | string | Log level | `info`, `error` |
| `subdomain` | string | Target subdomain | `test.example.com` |
| `status` | string | Check result | `vulnerable`, `http error` |
| `engine` | string | Service engine | `AWS S3`, `GitHub Pages` |
| `documentation` | string | Remediation URL | `https://...` |
| `discussion` | string | Discussion URL | `https://...` |
| `target_count` | int | Number of targets | `1000` |
| `fingerprint_count` | int | Fingerprints loaded | `44` |
| `concurrency` | int | Worker count | `10` |
| `timeout_seconds` | int | Request timeout | `10` |
| `output_file` | string | Result file | `results.json` |

---

## Troubleshooting

### Graylog Not Receiving Logs

**Check 1: Graylog is running**
```bash
docker-compose ps
# All services should be "Up"
```

**Check 2: Input is configured**
```bash
# In Graylog UI: System → Inputs
# Should see "GELF UDP" input running on port 12201
```

**Check 3: Port is accessible**
```bash
# Test UDP port
nc -u -v localhost 12201
```

**Check 4: Firewall rules**
```bash
# Allow UDP 12201
sudo ufw allow 12201/udp
```

### Logs Not Appearing in Graylog

**Check Subzy is sending logs**
```bash
# Run with debug logging
./subzy run --target example.com --graylog-host localhost:12201 --log-level debug

# Should not show connection errors
```

**Check Graylog logs**
```bash
docker-compose logs graylog | grep -i error
```

**Check Input statistics**
```bash
# In Graylog UI: System → Inputs → GELF UDP → Show received messages
# Should show incoming messages
```

### Connection Refused

```bash
# Error: "failed to create Graylog writer: connection refused"

# Solution 1: Check Graylog is running
docker-compose ps

# Solution 2: Check correct host/port
# Use: localhost:12201 (local)
# Not: graylog:12201 (Docker internal)
```

### No Data in Dashboards

```bash
# Check time range
# Graylog default: Last 5 minutes
# Change to: Last 24 hours

# Check search query
# Ensure: app:subzy (not app:"subzy")

# Check index
# System → Indices → Should see data in default index
```

---

## Performance Considerations

### GELF UDP Performance

- **UDP is fast** but can lose packets
- For production, consider **GELF TCP** (port 12201)
- UDP is fine for development and most use cases

### Resource Usage

**Graylog Stack Resources:**
- MongoDB: ~200MB RAM
- Elasticsearch: ~1GB RAM
- Graylog: ~1GB RAM
- Total: ~2.5GB RAM minimum

**Subzy Logging Overhead:**
- Console: Minimal
- GELF UDP: < 1ms per log
- File: < 5ms per log
- Total impact: < 5% performance

### Scaling Graylog

For high-volume logging:

1. **Increase Elasticsearch heap**:
   ```yaml
   ES_JAVA_OPTS: -Xms2g -Xmx2g
   ```

2. **Add Elasticsearch nodes** for clustering

3. **Use GELF TCP** for reliability

4. **Enable Graylog processing buffers**

---

## Integration Examples

### With CI/CD Pipeline

```yaml
# .gitlab-ci.yml
security_scan:
  script:
    - ./subzy run \
        --targets domains.txt \
        --graylog-host graylog.internal:12201 \
        --graylog-app "subzy-ci-${CI_PIPELINE_ID}" \
        --log-level info \
        --output results.json
    - cat results.json
```

### With Kubernetes

```yaml
# kubernetes/cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: subzy-scan
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: subzy
            image: subzy:latest
            args:
              - run
              - --targets
              - /config/domains.txt
              - --graylog-host
              - graylog-service.logging:12201
              - --graylog-app
              - subzy-k8s
              - --log-level
              - info
            volumeMounts:
            - name: domains
              mountPath: /config
```

### With Ansible

```yaml
# playbook.yml
- name: Run Subzy scan
  hosts: scanner
  tasks:
    - name: Run subdomain scan
      command:
        cmd: >
          /opt/subzy/subzy run
          --targets /opt/subzy/domains.txt
          --graylog-host {{ graylog_host }}:12201
          --graylog-app subzy-{{ inventory_hostname }}
          --log-level info
```

---

## Additional Resources

- [Graylog Documentation](https://docs.graylog.org/)
- [GELF Specification](https://docs.graylog.org/en/latest/pages/gelf.html)
- [Zerolog Documentation](https://github.com/rs/zerolog)
- [Subzy GitHub](https://github.com/LukaSikic/subzy)

---

## Support

For issues:
1. Check this guide
2. Review Graylog logs: `docker-compose logs graylog`
3. Open GitHub issue with logs and configuration

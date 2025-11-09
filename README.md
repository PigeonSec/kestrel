<div align="center">

![Kestrel Logo](docs/logo.png)

# Kestrel

**High-Performance MISP-Compliant Threat Intelligence Feed Server**

[![Go Version](https://img.shields.io/badge/Go-1.25.4+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Custom-blue.svg)](LICENSE)

Lightweight, self-hosted IOC feed distribution engine. Serves structured MISP-compatible JSON feeds and TXT-formatted blocklists for Pi-hole, firewalls, proxies, and DNS sinkhole systems.

**Designed for speed, simplicity, and data-source agnosticism** — Kestrel runs entirely on your own infrastructure.

[Features](#features) • [Quick Start](#quick-start) • [Use Cases](#use-cases) • [API](#api-endpoints) • [Testing](TESTING.md)

</div>

---

## Overview

Kestrel is a self-hosted threat intelligence distribution platform that makes sharing IOCs (Indicators of Compromise) simple and efficient. Whether you're running a security team, managing enterprise DNS filtering, or building custom threat feeds, Kestrel provides:

- **MISP-compliant JSON feeds** for threat intelligence platforms
- **TXT blocklists** for Pi-hole, AdGuard, firewalls, and DNS sinkholes
- **Flexible storage** with in-memory caching and Redis/Valkey backends
- **Domain validation** to ensure IOCs are active and reachable
- **API key management** with SQLite persistence and optional external sync
- **Simple deployment** as a single static Go binary

## Features

- ⚡ **Blazing Fast** - Built with Go, Valkey/Redis, and in-memory caching
- 🧠 **Flexible Storage** - In-memory (dev/testing) or Valkey/Redis (production)
- 🔐 **API Key Management** - Built-in generation with `kestrel_` prefix
- 🌐 **Agnostic Auth** - Sync keys from any external API or use local SQLite
- ✅ **Domain Validation** - DNS (A/AAAA/CNAME) and HTTP/HTTPS connectivity checks
- 📊 **MISP Compliant** - Standard event format, manifest, and attributes
- 📜 **Blocklists** - Generate TXT feeds for Pi-hole, AdGuard, firewalls
- 🐳 **Docker Ready** - Compose setup with Valkey included
- 🧱 **Simple Deployment** - Single binary + Redis/Valkey + optional SQLite

## Use Cases

### 🛡️ Threat Intelligence Sharing
Distribute IOCs across your security infrastructure using standard MISP feeds. Kestrel acts as a central distribution point for threat intelligence, allowing security tools to consume feeds programmatically.

### 🏢 Enterprise DNS Filtering
Deploy organization-wide DNS blocklists to Pi-hole, AdGuard, or custom DNS resolvers. Update blocklists in real-time as new threats are identified, protecting all devices on your network.

### 🔍 SIEM Integration
Feed threat indicators directly into your SIEM platform via MISP-compatible JSON. Correlate IOCs with logs and alerts to identify compromised systems faster.

### 🔬 Security Research
Host your own private threat intelligence feeds for research purposes. Validate domains before adding them to ensure accuracy and reduce false positives.

### 🌐 Custom Threat Feeds
Build specialized feeds for specific threat types, geographic regions, or industry sectors. Support both free (public) and paid (premium) access models.

## Quick Start

```bash
# 1. Build
go build -o kestrel ./cmd/kestrel

# 2. Generate API key
./kestrel -generate-key
# Output: Generated API key: kestrel_abc123...

# 3. Add key to database
sqlite3 kestrel.db "INSERT INTO accounts VALUES ('kestrel_YOUR_KEY', 'admin@example.com', 'admin', 1);"

# 4. Start server
STORAGE_TYPE=memory ./kestrel

# 5. Test
curl http://localhost:8080/healthz
```

## Configuration

See [.env.example](.env.example) for all options:

```bash
STORAGE_TYPE=memory           # memory (dev) or valkey (prod)
VALKEY_ADDR=localhost:6379
LISTEN_ADDR=:8080
ENABLE_VALIDATION=true        # Validate domains before adding
```

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/healthz` | No | Health check |
| `GET` | `/pihole/public.txt` | No | Public blocklist (free) |
| `POST` | `/api/ioc` | Yes | Submit IOC (supports `?validate=dns\|http\|full`) |
| `GET` | `/misp/manifest.json` | Yes | MISP manifest |
| `GET` | `/misp/events/:id.json` | Yes | MISP event details |
| `GET` | `/pihole/:feed.txt` | Yes | Premium blocklist (paid) |
| `POST` | `/api/admin/generate-key` | Yes | Generate new API key |
| `GET` | `/api/admin/accounts` | Yes | List all accounts |

**Authentication**: Use `X-API-Key` header, `Authorization: Bearer` header, or `?apikey=` query param.

**Validation Modes**:
- `?validate=dns` - Check for A, AAAA, or CNAME records
- `?validate=http` - Verify HTTP/HTTPS connectivity
- `?validate=full` - Both DNS and HTTP validation

## Examples

### Submit IOC with DNS Validation
```bash
curl -X POST http://localhost:8080/api/ioc?validate=dns \
  -H "X-API-Key: kestrel_YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "malicious.com",
    "category": "Network activity",
    "comment": "C2 server detected via sandbox",
    "feed": "premium"
  }'
```

### Consume MISP Feed
```bash
# Get manifest
curl http://localhost:8080/misp/manifest.json \
  -H "X-API-Key: kestrel_YOUR_KEY"

# Get specific event
curl http://localhost:8080/misp/events/<event-id>.json \
  -H "X-API-Key: kestrel_YOUR_KEY"
```

### Use with Pi-hole
```bash
# Add to Pi-hole Adlists:
# Public feed (no auth)
http://localhost:8080/pihole/public.txt

# Premium feed (requires auth)
http://localhost:8080/pihole/premium.txt?apikey=kestrel_YOUR_KEY
```

### Firewall / DNS Sinkhole Integration
```bash
# Periodic fetch for firewall rules
curl -s http://localhost:8080/pihole/premium.txt?apikey=YOUR_KEY > /etc/blocklist.txt
```

## Docker Deployment

```bash
# Start with Valkey
docker-compose up -d

# Generate admin key
docker exec kestrel ./kestrel -generate-key

# Add key to database
docker exec kestrel sqlite3 /data/kestrel.db \
  "INSERT INTO accounts VALUES ('YOUR_KEY', 'admin@example.com', 'admin', 1);"
```

## Architecture

```
+-----------------------+
| External Key Source   |  ← Optional: any HTTPS API for subscriber management
+----------+------------+
           |
           v
+----------v------------+
|     Kestrel (Go)      |
|  - HTTP/HTTPS         |
|  - Valkey/Redis       |
|  - SQLite fallback    |
+----------+------------+
           |
   ┌───────┼─────────────────────┐
   │       │                     │
   v       v                     v
MISP JSON Feeds     TXT Blocklists     API Ingestion
(/misp/...)         (/pihole/...)      (POST /api/ioc)
```

## Performance

- **In-memory event caching** for instant MISP feed delivery
- **Concurrent request handling** via Gin framework
- **Valkey/Redis** for distributed deployments
- **SQLite** for persistent API key storage with zero config
- **Minimal allocations** in hot code paths

## Testing

```bash
# Basic compliance tests
./test/test_kestrel.sh

# Authenticated API tests
KESTREL_API_KEY=your-key ./test/test_with_auth_simple.sh
```

See [TESTING.md](TESTING.md) for comprehensive test documentation and MISP compliance verification.

## Project Structure

```
kestrel/
├── cmd/kestrel/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── storage/         # Storage backends (memory, valkey)
│   ├── auth/            # Authentication & key management
│   ├── validation/      # Domain validation (DNS, HTTP)
│   ├── misp/            # MISP event handling
│   └── handlers/        # HTTP request handlers
├── test/                # Test scripts
├── .env.example         # Configuration template
├── docker-compose.yml   # Docker setup
└── Dockerfile           # Container image
```

## CLI Flags

```bash
./kestrel -help           # Show help
./kestrel -version        # Show version
./kestrel -generate-key   # Generate API key
```

## License

Custom Attribution License - See [LICENSE](LICENSE)

**TL;DR**: Free to use with attribution required. Include "Powered by Kestrel by PigeonSec" or link to this repo.

---

<div align="center">

Made by [Karl Machleidt](https://github.com/pigeonsec) / **PigeonSec**

⭐ Star this repo if you find it useful!

[Report Issues](https://github.com/pigeonsec/kestrel/issues) • [View License](LICENSE) • [Read Docs](TESTING.md)

</div>

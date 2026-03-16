# vigil

[![Docker Image](https://img.shields.io/badge/docker-ghcr.io%2Fmeysam81%2Fvigil-blue)](https://github.com/meysam81/vigil/pkgs/container/vigil)
[![Go Report Card](https://goreportcard.com/badge/github.com/meysam81/vigil)](https://goreportcard.com/report/github.com/meysam81/vigil)
[![License](https://img.shields.io/github/license/meysam81/vigil)](LICENSE)

Lightweight service to collect and persist Content Security Policy (CSP) violation reports in Redis for audit and investigation.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Quick Start](#quick-start)
- [Configure Your CSP Header](#configure-your-csp-header)
  - [Legacy Format (report-uri)](#legacy-format-report-uri)
  - [Modern Format (report-to)](#modern-format-report-to)
  - [Hybrid Approach (Recommended)](#hybrid-approach-recommended)
- [Configuration](#configuration)
- [Data Storage](#data-storage)
- [Report Formats](#report-formats)
  - [Legacy Report-URI Format](#legacy-report-uri-format)
  - [Modern Reporting API Format](#modern-reporting-api-format)
- [Docker Compose Example](#docker-compose-example)
- [Rate Limiting](#rate-limiting)
- [License](#license)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Quick Start

```bash
docker run --rm -dp 8080:8080 \
  -e REDIS_HOST=your-redis-host \
  ghcr.io/meysam81/vigil
```

## Configure Your CSP Header

The collector supports both legacy and modern CSP reporting formats:

### Legacy Format (report-uri)

For older browsers and simpler setups:

```shell
Content-Security-Policy: default-src 'self'; report-uri https://csp.example.com
```

This sends reports with `Content-Type: application/csp-report` or `application/json`.

### Modern Format (report-to)

For modern browsers with enhanced reporting capabilities:

```shell
Content-Security-Policy: default-src 'self'; report-to csp-endpoint
Report-To: {"group":"csp-endpoint","max_age":86400,"endpoints":[{"url":"https://csp.example.com"}]}
```

This sends reports with `Content-Type: application/reports+json`.

#### Alternative: Reporting Endpoints

You can also use the newer `Reporting-Endpoints` header instead of `Report-To`:

```shell
Content-Security-Policy: default-src 'self'; report-to csp-endpoint
Reporting-Endpoints: csp-endpoint="https://csp.example.com"
```

The `Reporting-Endpoints` header provides a simpler syntax compared to `Report-To` and is supported by modern browsers.

### Hybrid Approach (Recommended)

For maximum compatibility, use all the headers to support old and new browsers:

```shell
Content-Security-Policy: default-src 'self'; report-uri https://csp.example.com; report-to csp-endpoint
Report-To: {"group":"csp-endpoint","max_age":86400,"endpoints":[{"url":"https://csp.example.com"}]}
Reporting-Endpoints: csp-endpoint="https://csp.example.com"
```

## Configuration

All configuration follows the 12-factor app methodology via environment variables:

| Variable             | Default     | Description                           |
| -------------------- | ----------- | ------------------------------------- |
| `PORT`               | `8080`      | HTTP server port                      |
| `LOG_LEVEL`          | `info`      | Log verbosity (debug/info/warn/error) |
| `REDIS_HOST`         | `localhost` | Redis server hostname (**required**)  |
| `REDIS_PORT`         | `6379`      | Redis server port                     |
| `REDIS_DB`           | `0`         | Redis database number                 |
| `REDIS_PASSWORD`     | -           | Redis authentication password         |
| `REDIS_SSL__ENABLED` | `false`     | Enable TLS connection to Redis        |
| `RATELIMIT_MAX`      | `20`        | Max requests per IP                   |
| `RATELIMIT_REFILL`   | `2.0`       | Token refill rate per second          |

## Data Storage

CSP reports are stored in Redis with:

- **Key**: Unix timestamp of receipt
- **Value**: Full JSON report
- **TTL**: Indefinite (configure Redis eviction policy as needed)

## Report Formats

The collector accepts both CSP reporting formats:

### Legacy Report-URI Format

```json
{
  "csp-report": {
    "blocked-uri": "inline",
    "document-uri": "https://example.com/page",
    "effective-directive": "script-src-elem",
    "original-policy": "default-src 'self'",
    "referrer": "https://example.com/",
    "status-code": 200,
    "violated-directive": "script-src-elem"
  }
}
```

### Modern Reporting API Format

```json
{
  "age": 53531,
  "body": {
    "blockedURL": "inline",
    "disposition": "enforce",
    "documentURL": "https://example.com/page",
    "effectiveDirective": "script-src-elem",
    "originalPolicy": "default-src 'self'",
    "statusCode": 200
  },
  "type": "csp-violation",
  "url": "https://example.com/page",
  "user_agent": "Mozilla/5.0..."
}
```

## Docker Compose Example

```yaml
version: "3.8"
services:
  csp-collector:
    image: ghcr.io/meysam81/vigil
    ports:
      - "8080:8080"
    environment:
      REDIS_HOST: redis
      RATELIMIT_MAX: 50
    depends_on:
      - redis

  redis:
    image: redis:alpine
    volumes:
      - redis-data:/data

volumes:
  redis-data:
```

## Rate Limiting

Built-in rate limiting per IP address returns:

- `X-RateLimit-Total` header with limit
- `X-RateLimit-Remaining` header with remaining requests
- `429 Too Many Requests` when exceeded

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

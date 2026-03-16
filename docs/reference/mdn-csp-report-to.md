# CSP report-to Directive -- MDN Reference

> Source: <https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/report-to>
> Spec Status: Living Standard (CSP Level 3)
> Last Fetched: 2026-03-16

## Overview

The `report-to` directive instructs the user agent to send CSP violation reports to a named endpoint group defined by the `Reporting-Endpoints` HTTP response header. It is the modern replacement for the deprecated `report-uri` directive and uses the Reporting API to deliver structured violation reports. The directive has no effect in and of itself, but only gains meaning in combination with other directives.

CSP Version: 3
Directive Type: Reporting directive
Meta Element Support: Not supported in the `<meta>` element.

## Syntax

```
Content-Security-Policy: ...; report-to <endpoint_name>
```

`<endpoint_name>` -- The name of an endpoint defined in the `Reporting-Endpoints` (preferred) or the deprecated `Report-To` HTTP response header. The browser resolves this name to a URL and sends violation reports there.

## Interaction with Reporting-Endpoints Header

The `report-to` directive references an endpoint name that must be defined in a separate HTTP response header. The server must provide the mapping between endpoint names and URLs using the `Reporting-Endpoints` header. Both headers must be present in the same HTTP response for reporting to work.

```
Reporting-Endpoints: <name>="<URL>"
```

The `Reporting-Endpoints` header maps a short name to a full URL. The CSP `report-to` directive then references that name. Without the corresponding `Reporting-Endpoints` header, the browser has no URL to send reports to and will silently discard them.

## Example Headers

A complete working configuration requires both headers:

```http
Reporting-Endpoints: csp-endpoint="https://example.com/csp-reports"
Content-Security-Policy: default-src 'self'; report-to csp-endpoint
```

Multiple endpoints can be defined in a single `Reporting-Endpoints` header:

```http
Reporting-Endpoints: csp-endpoint="https://example.com/csp-reports", default="https://example.com/reports"
Content-Security-Policy: default-src 'self'; report-to csp-endpoint
```

## Report Format

Reports are sent via HTTP POST with the Content-Type:

```
Content-Type: application/reports+json
```

Each report is a JSON-serialized `Report` object (per the Reporting API specification). For CSP violations, the `type` is `"csp-violation"` and the `body` contains a serialized `CSPViolationReportBody` object.

### Report Envelope Fields

| Field        | Type   | Description                                              |
| ------------ | ------ | -------------------------------------------------------- |
| `age`        | number | Age of the report in milliseconds since generation.      |
| `type`       | string | Always `"csp-violation"` for CSP reports.                |
| `url`        | string | The URL of the document where the violation occurred.    |
| `user_agent` | string | The User-Agent string of the reporting browser.          |
| `body`       | object | The `CSPViolationReportBody` object (see fields below).  |

### CSPViolationReportBody Fields

| Field                 | Type    | Description                                                                                                       |
| --------------------- | ------- | ----------------------------------------------------------------------------------------------------------------- |
| `blockedURL`          | string  | The URL of the resource that was blocked by CSP.                                                                  |
| `columnNumber`        | integer | The column number in the source file where the violation occurred.                                                |
| `disposition`         | string  | Either `"enforce"` or `"report"` depending on which CSP header triggered the report.                              |
| `documentURL`        | string  | The URL of the document in which the violation occurred.                                                          |
| `effectiveDirective`  | string  | The specific directive whose enforcement caused the violation (e.g., `script-src-elem`, `style-src-elem`).        |
| `lineNumber`          | integer | The line number in the source file where the violation occurred.                                                  |
| `originalPolicy`      | string  | The original policy as specified in the `Content-Security-Policy` header.                                         |
| `referrer`            | string  | The referrer of the document in which the violation occurred.                                                     |
| `sample`              | string  | The first 40 characters of the inline script, event handler, or style that caused the violation.                  |
| `sourceFile`          | string  | The URL of the file in which the violation was triggered.                                                         |
| `statusCode`          | integer | The HTTP status code of the resource on which the global object was instantiated.                                 |

### Example Report Body

From the MDN documentation, a complete CSP violation report as delivered to the endpoint:

```json
{
  "age": 53531,
  "body": {
    "blockedURL": "inline",
    "columnNumber": 39,
    "disposition": "enforce",
    "documentURL": "https://example.com/csp-report",
    "effectiveDirective": "script-src-elem",
    "lineNumber": 121,
    "originalPolicy": "default-src 'self'; report-to csp-endpoint-name",
    "referrer": "https://www.google.com/",
    "sample": "console.log(\"lo\")",
    "sourceFile": "https://example.com/csp-report",
    "statusCode": 200
  },
  "type": "csp-violation",
  "url": "https://example.com/csp-report",
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36"
}
```

## Migration from report-uri

The `report-to` directive is intended to replace `report-uri`. During the transition period, sites should specify both directives for backward compatibility. Browsers that support `report-to` will ignore the `report-uri` directive when both are present; browsers that do not support `report-to` will fall back to `report-uri`.

### Example Migration Header

```http
Content-Security-Policy: default-src 'self'; report-uri https://endpoint.example.com/csp; report-to csp-endpoint
```

### Key Differences Between Formats

| Aspect             | `report-uri` (Legacy)                | `report-to` (Modern)                             |
| ------------------ | ------------------------------------ | ------------------------------------------------- |
| Content-Type       | `application/csp-report`             | `application/reports+json`                        |
| JSON Envelope      | `{"csp-report": {...}}`              | `{"type": "csp-violation", "body": {...}, ...}`   |
| Field Naming       | kebab-case (`blocked-uri`)           | camelCase (`blockedURL`)                          |
| Endpoint Config    | Inline URL in directive              | Named endpoint via `Reporting-Endpoints` header   |
| Sample Field       | `script-sample`                      | `sample`                                          |
| Extra Envelope     | None (just `csp-report` wrapper)     | `age`, `type`, `url`, `user_agent` at root level  |

## Browser Compatibility

The MDN browser compatibility table was not available in rendered form at fetch time. MDN marks this feature as:

> **Limited availability** -- This feature is not Baseline because it does not work in some of the most widely-used browsers.

In general:

- **Chrome/Edge:** Support `report-to` (Chromium-based browsers were early adopters of the Reporting API).
- **Firefox:** Has historically not supported the `report-to` directive or the Reporting API endpoints.
- **Safari:** Has historically not supported `report-to`.
- Both `report-uri` and `report-to` should be specified simultaneously until `report-to` reaches broader adoption.

Consult the [MDN compatibility table](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/report-to#browser_compatibility) for current data.

## Specification

[W3C Content Security Policy Level 3 -- directive-report-to](https://w3c.github.io/webappsec-csp/#directive-report-to)

## Relevance to Vigil

- Vigil's modern endpoint receives reports via the Reporting API format with `Content-Type: application/reports+json`.
- The report body uses **camelCase** field names (`CSPViolationReportBody`) -- distinct from the legacy kebab-case `csp-report` format.
- The modern report envelope includes metadata (`age`, `type`, `url`, `user_agent`) at the root level alongside `body`.
- Vigil must handle both `report-uri` and `report-to` formats simultaneously for full browser coverage during the transition period.
- The `disposition` field (`"enforce"` or `"report"`) is present in both formats and useful for filtering in Vigil's Slack aggregate reporter.
- The `blockedURL` field may contain the string `"inline"` (for inline script/style violations) or `"eval"` (for eval violations) rather than actual URLs.
- Security note: all report data is potentially attacker-controlled and must be treated as untrusted input -- sanitize before storing or rendering.
- Vigil stores raw report JSON in Redis; for modern reports, the camelCase field names are what appear in storage.

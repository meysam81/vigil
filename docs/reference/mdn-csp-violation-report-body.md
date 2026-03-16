# CSPViolationReportBody — MDN Web API Reference

> Source: <https://developer.mozilla.org/en-US/docs/Web/API/CSPViolationReportBody>
> Spec: [Content Security Policy Level 3 — CSPViolationReportBody](https://w3c.github.io/webappsec-csp/#dictdef-cspviolationreportbody)
> Spec Status: Living Standard
> Last Fetched: 2026-03-16

## Overview

`CSPViolationReportBody` is the Reporting API interface that represents the body of a Content Security Policy (CSP) violation report. CSP violations occur when a webpage attempts to load a resource that violates the policy set by the `Content-Security-Policy` HTTP header; reports are delivered either to a `ReportingObserver` callback (with `type: "csp-violation"`) or as JSON payloads to endpoints declared via the `report-to` directive.

## Inheritance

```
ReportBody
  └── CSPViolationReportBody
```

`CSPViolationReportBody` extends `ReportBody`. Available only in secure contexts (HTTPS).

## Properties

All properties are **read-only** instance properties.

| Property             | Type                | Description |
|----------------------|---------------------|-------------|
| `blockedURL`         | `string`            | Either the type or the URL of the resource that was blocked because it violates the CSP. |
| `columnNumber`       | `number` or `null`  | The column number in the script at which the violation occurred. |
| `disposition`        | `string`            | How the violated policy is configured to be treated by the user agent. Either `"enforce"` or `"report"`. |
| `documentURL`        | `string`            | The URL of the document or worker in which the violation was found. |
| `effectiveDirective` | `string`            | The directive whose enforcement uncovered the violation (e.g. `script-src-elem`). |
| `lineNumber`         | `number` or `null`  | The line number in the script at which the violation occurred. |
| `originalPolicy`     | `string`            | The policy whose enforcement uncovered the violation (the full CSP policy string). |
| `referrer`           | `string` or `null`  | The URL for the referrer of the resource whose policy was violated, or `null`. |
| `sample`             | `string`            | The first 40 characters of the inline resource that caused the violation (if the resource is an inline script, event handler, or style and `'report-sample'` is present in the directive). External resources do not generate a sample. |
| `sourceFile`         | `string` or `null`  | The URL of the script where the violation occurred; `null` if the violation did not occur in a script. Both `columnNumber` and `lineNumber` should be non-null when this is non-null. |
| `statusCode`         | `number`            | The HTTP status code of the document or worker in which the violation occurred. |

### Methods

- `toJSON()` — *(deprecated)* Returns a JSON representation of the `CSPViolationReportBody` object.

## JSON Serialization

When reports are sent to a reporting endpoint via `report-to`, they arrive as a JSON array. Each element has top-level metadata (`age`, `type`, `url`, `user_agent`) and a `body` object whose fields are the camelCase property names listed above.

### Example: Reporting API payload (sent to endpoint)

```json
[
  {
    "age": 53531,
    "body": {
      "blockedURL": "inline",
      "columnNumber": 59,
      "disposition": "enforce",
      "documentURL": "https://example.com/csp-report",
      "effectiveDirective": "script-src-elem",
      "lineNumber": 1441,
      "originalPolicy": "default-src 'self'; report-to csp-endpoint",
      "referrer": "https://www.google.com/",
      "sample": "",
      "sourceFile": "https://example.com/csp-report",
      "statusCode": 200
    },
    "type": "csp-violation",
    "url": "https://example.com/csp-report",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36"
  }
]
```

### Example: ReportingObserver output (in-browser)

```json
{
  "type": "csp-violation",
  "url": "http://127.0.0.1:9999/",
  "body": {
    "sourceFile": null,
    "lineNumber": null,
    "columnNumber": null,
    "documentURL": "http://127.0.0.1:9999/",
    "referrer": "",
    "blockedURL": "https://apis.google.com/js/platform.js",
    "effectiveDirective": "script-src-elem",
    "originalPolicy": "default-src 'self';",
    "sample": "",
    "disposition": "enforce",
    "statusCode": 200
  }
}
```

## Differences from Legacy `csp-report` Format

The legacy `report-uri` directive sends reports wrapped in `{"csp-report": {...}}` using **kebab-case** field names. The modern `report-to` directive sends reports wrapped in `[{"body": {...}, ...}]` using **camelCase** field names.

| Modern (`body.*`)    | Legacy (`csp-report.*`)        | Notes |
|----------------------|-------------------------------|-------|
| `blockedURL`         | `blocked-uri`                 | Modern uses "URL", legacy uses "uri" |
| `columnNumber`       | `column-number`               | Optional in both; CSP Level 2 |
| `disposition`        | `disposition`                 | Present in both formats |
| `documentURL`        | `document-uri`                | Modern uses "URL", legacy uses "uri" |
| `effectiveDirective` | `effective-directive`         | Same semantics |
| `lineNumber`         | `line-number`                 | Optional in both; CSP Level 2 |
| `originalPolicy`     | `original-policy`             | Same semantics |
| `referrer`           | `referrer`                    | Identical in both |
| `sample`             | `script-sample`               | Different field name; both cap at 40 chars |
| `sourceFile`         | `source-file`                 | Optional in both; CSP Level 2 |
| `statusCode`         | `status-code`                 | Same semantics |
| *(none)*             | `violated-directive`          | Legacy-only; historic alias for `effective-directive` |

### Structural differences

| Aspect | Legacy (`report-uri`) | Modern (`report-to`) |
|--------|----------------------|---------------------|
| Wrapper | `{"csp-report": {...}}` | `[{"body": {...}, "type": "csp-violation", ...}]` |
| Content-Type | `application/csp-report` | `application/reports+json` |
| Top-level metadata | None | `age`, `type`, `url`, `user_agent` |
| Field casing | kebab-case | camelCase |
| Directive | `report-uri <url>` | `report-to <group-name>` + `Reporting-Endpoints` header |

## Browser Compatibility

CSPViolationReportBody requires a **secure context** (HTTPS). The Reporting API (`report-to`) has broad Chromium support but historically lagged in Firefox and Safari. For precise version numbers, consult the [MDN compatibility table](https://developer.mozilla.org/en-US/docs/Web/API/CSPViolationReportBody#browser_compatibility) or [Can I Use](https://caniuse.com/mdn-api_cspviolationreportbody).

General guidance (as of 2026-03-16):

- **Chrome / Edge**: Supported since early Chromium adoption of the Reporting API.
- **Firefox**: Added support for `report-to` and CSPViolationReportBody in later releases.
- **Safari**: Added support in recent versions.
- **Recommendation**: Use both `report-uri` and `report-to` simultaneously for maximum coverage until legacy support is fully phased out.

## Relevance to Vigil

- `CSPViolationReportBody` defines **all fields** Vigil may encounter in modern (`report-to`) reports. The `body` object in the JSON payload uses these exact camelCase names.
- Vigil's gjson field extraction paths in the Slack aggregate reporter (`internal/reporter/aggregate.go`) must match these exact field names:
  - `body.effectiveDirective`
  - `body.blockedURL`
  - `body.documentURL`
  - `user_agent` (top-level, not inside `body`)
- For legacy `report-uri` payloads, Vigil uses the corresponding kebab-case paths under `csp-report`:
  - `csp-report.effective-directive`
  - `csp-report.blocked-uri`
  - `csp-report.document-uri`
- The `disposition` field tells Vigil whether the report came from an enforced policy (`"enforce"`) or a report-only policy (`"report"`), which can be used to prioritize alerts.
- Vigil stores raw JSON bodies without model structs, so understanding both field-naming conventions is critical for any future gjson queries or Slack message formatting.

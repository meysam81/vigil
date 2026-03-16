# CSP report-uri Directive -- MDN Reference

> Source: <https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/report-uri>
> Spec Status: Deprecated (use report-to instead)
> Last Fetched: 2026-03-16

## Overview

The `report-uri` directive instructs the user agent to report attempts to violate the Content Security Policy. These violation reports consist of JSON documents sent via an HTTP POST request to the specified URI. The directive has no effect in and of itself, but only gains meaning in combination with other directives.

## Deprecation Status

The `report-to` directive is intended to replace `report-uri`. In browsers that support `report-to`, the `report-uri` directive is ignored. However, until `report-to` is broadly supported, sites should specify both directives for backward compatibility:

```
Content-Security-Policy: ...; report-uri https://endpoint.example.com; report-to endpoint_name
```

CSP Version: 1
Directive Type: Reporting directive
Meta Element Support: Not supported in the `<meta>` element.

## Syntax

```
Content-Security-Policy: report-uri <uri>;
Content-Security-Policy: report-uri <uri> <uri>;
```

`<uri>` -- A URI indicating where the report must be sent. Multiple URIs can be specified, separated by spaces.

## Legacy Report Format

Reports are sent as JSON via HTTP POST. The body is a single JSON object with a top-level `"csp-report"` key wrapping all violation fields.

### csp-report Wrapper

The report is wrapped in a `{"csp-report": {...}}` JSON object. This is distinct from the modern Reporting API format, which uses a different envelope structure.

### Complete Field List

| Field                  | Type    | Description                                                                                                                                                                           |
| ---------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `document-uri`         | string  | The URI of the document in which the violation occurred.                                                                                                                              |
| `referrer`             | string  | **(Deprecated, Non-standard)** The referrer of the document in which the violation occurred.                                                                                          |
| `violated-directive`   | string  | **(Deprecated)** Historic name for `effective-directive`; contains the same value.                                                                                                    |
| `effective-directive`  | string  | The directive whose enforcement caused the violation. Some browsers may provide different values (e.g., Chrome provides `style-src-elem`/`style-src-attr` even when `style-src` was enforced). |
| `original-policy`      | string  | The original policy as specified by the `Content-Security-Policy` HTTP header.                                                                                                        |
| `blocked-uri`          | string  | The URI of the resource blocked from loading by CSP. If from a different origin than `document-uri`, truncated to scheme, host, and port only.                                        |
| `status-code`          | number  | The HTTP status code of the resource on which the global object was instantiated.                                                                                                     |
| `disposition`          | string  | Either `"enforce"` or `"report"` depending on whether `Content-Security-Policy` or `Content-Security-Policy-Report-Only` header is used.                                              |
| `source-file`          | string  | The URI of the document or worker in which the violation was triggered. Not explicitly listed on MDN but present in real-world reports.                                                |
| `line-number`          | integer | The line number in the source file at which the violation occurred. Not explicitly listed on MDN but present in real-world reports.                                                    |
| `column-number`        | integer | The column number in the source file at which the violation occurred. Not explicitly listed on MDN but present in real-world reports.                                                  |
| `script-sample`        | string  | The first 40 characters of the inline script, event handler, or style that caused the violation. Only for `script-src*` and `style-src*` violations when the directive contains `'report-sample'`. Violations from external files are not included. |

### Example JSON

A page at `http://example.com/signup.html` with the policy:

```
Content-Security-Policy: default-src 'none'; style-src cdn.example.com; report-uri /_/csp-reports
```

Attempts to load CSS from `http://example.com/css/style.css` (violating the policy). The following report is sent to `http://example.com/_/csp-reports`:

```json
{
  "csp-report": {
    "blocked-uri": "http://example.com/css/style.css",
    "disposition": "report",
    "document-uri": "http://example.com/signup.html",
    "effective-directive": "style-src-elem",
    "original-policy": "default-src 'none'; style-src cdn.example.com; report-uri /_/csp-reports",
    "referrer": "",
    "status-code": 200,
    "violated-directive": "style-src-elem"
  }
}
```

## Content-Type

Reports are sent as HTTP POST requests with the header:

```
Content-Type: application/csp-report
```

## Blocked URI Truncation Rules

- **Same-origin resources:** The full path is included in `blocked-uri`.
- **Cross-origin resources:** Truncated to **scheme, host, and port** only (no path) to prevent leaking sensitive information about cross-origin resources.

Example: if a page attempts to load `http://anothercdn.example.com/stylesheet.css`, the report includes only `http://anothercdn.example.com` as the `blocked-uri`.

Reference: [W3C CSP specification -- Security violation reports](https://w3c.github.io/webappsec-csp/#security-violation-reports)

## script-sample Behavior

- **Scope:** Only applicable to `script-src*` and `style-src*` violations.
- **Requirement:** The corresponding CSP directive must contain the `'report-sample'` keyword.
- **Content:** First 40 characters of the inline script, event handler, or style that caused the violation.
- **Exclusion:** Violations originating from external files are NOT included.

## Security Considerations

Per MDN:

> Violation reports should be considered attacker-controlled data. The content should be properly sanitized before storing or rendering. This is particularly true of the `script-sample` property, if supplied.

Additionally:

- `blocked-uri`, `source-file`, `document-uri`, and `referrer` can all contain attacker-influenced values.
- Cross-origin URI truncation mitigates some information leakage, but same-origin URIs include full paths.
- Report endpoints should validate and sanitize all fields before storage or display.

## Browser Compatibility

The MDN browser compatibility table was not available in rendered form at fetch time. In general:

- `report-uri` is widely supported across all major browsers (Chrome, Firefox, Safari, Edge) due to its long history in CSP Level 1.
- In browsers that support `report-to`, the `report-uri` directive is ignored when both are present.
- Safari has historically lagged on `report-to` support, making `report-uri` still necessary for full cross-browser coverage.

Consult the [MDN compatibility table](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/report-uri#browser_compatibility) for current data.

## Relevance to Vigil

- Vigil's legacy endpoint (`/csp-report` or similar) accepts `application/csp-report` POST requests.
- The `csp-report` JSON wrapper is what Vigil parses at the legacy endpoint.
- All field names use **kebab-case** (hyphenated), unlike the modern camelCase Reporting API format.
- Vigil stores raw report JSON in Redis -- these are the field names present in legacy reports.
- The `disposition` field indicates whether the policy was enforced or report-only, which is useful for filtering in Vigil's Slack aggregate reporter.
- The `violated-directive` and `effective-directive` fields contain the same value; both may appear in reports.
- Security note: all report data is potentially attacker-controlled and must be treated as untrusted input.

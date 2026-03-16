# CSP Level 2 — W3C Recommendation

> Source: <https://www.w3.org/TR/CSP2/>
> Spec Status: W3C Recommendation (15 December 2016)
> Last Fetched: 2026-03-16

## Overview

Content Security Policy Level 2 defines a policy language used to declare a set of content restrictions for a web resource, and a mechanism for transmitting the policy from a server to a client where the policy is enforced. CSP2 is the first standardized version of CSP to include a fully specified violation reporting mechanism via the `report-uri` directive.

## Legacy Report Format (csp-report)

When a violation is detected, the browser constructs a JSON object with a single top-level key `csp-report` whose value is an object containing the violation details:

```json
{
  "csp-report": {
    "document-uri": "https://example.com/page",
    "referrer": "https://example.com/",
    "violated-directive": "default-src 'self'",
    "effective-directive": "script-src",
    "original-policy": "default-src 'self'; report-uri /csp-report",
    "blocked-uri": "https://evil.example.com/script.js",
    "status-code": 200
  }
}
```

### Field Definitions

All fields below are part of the violation report object nested under the `csp-report` key.

| Field | Required | Description |
|---|---|---|
| `document-uri` | Yes | The address of the protected resource, stripped for reporting (see Strip URI algorithm below). |
| `referrer` | Yes | The referrer attribute of the protected resource, or the empty string if the protected resource has no referrer. |
| `violated-directive` | Yes | The policy directive that was violated, as it appears in the policy. This will contain the `default-src` directive in the case of violations caused by falling back to the default sources when enforcing a directive. |
| `effective-directive` | Yes | The name of the policy directive that was violated. This will contain the directive whose enforcement triggered the violation (e.g. `script-src`) even if that directive does not explicitly appear in the policy, but is implicitly activated via the `default-src` directive. |
| `original-policy` | Yes | The original policy, as received by the user agent. |
| `blocked-uri` | Yes | The originally requested URL of the resource that was prevented from loading, stripped for reporting, or the empty string if the resource has no URL (inline script and inline style, for example). |
| `status-code` | Yes | The `status-code` of the HTTP response that contained the protected resource, if the protected resource was obtained over HTTP. Otherwise, the number `0`. |
| `source-file` | Optional | The URL of the resource where the violation occurred, stripped for reporting. Included when a specific file can be identified as the cause of the violation. |
| `line-number` | Optional | The line number in `source-file` on which the violation occurred. Included when a specific line can be identified. |
| `column-number` | Optional | The column number in `source-file` on which the violation occurred. Included when a specific column can be identified. |

**Note:** CSP Level 2 does NOT define the `script-sample` field. That field (first 40 characters of the inline script/style/event handler) was introduced in CSP Level 3 alongside the `'report-sample'` source expression.

**Note:** CSP Level 2 does NOT define the `disposition` field (which indicates `"enforce"` vs `"report"`). That field was also introduced in CSP Level 3.

### Content-Type

The report is sent as an HTTP `POST` request with:

```
Content-Type: application/csp-report
```

## Send Violation Reports Algorithm (Section 4.4)

The spec defines the following algorithm for sending violation reports:

1. Prepare a JSON object (`report object`) with a single key, `csp-report`, whose value is the result of generating a violation report object (the fields listed above).

2. Let `report body` be the JSON stringification of `report object`.

3. For each `report URL` in the set of report URLs (from the `report-uri` directive):
   - **Deduplication (MAY):** If the user agent has already sent a violation report for the protected resource to `report URL`, and that report contained an entity body that exactly matches `report body`, the user agent MAY abort these steps and continue to the next report URL.
   - **Send:** Queue a task to fetch `report URL` from the origin of the protected resource, with the synchronous flag not set, using HTTP method `POST`, with a `Content-Type` header field of `application/csp-report`, and an entity body consisting of `report body`.
   - **Cross-origin cookies:** If the origin of `report URL` is not the same as the origin of the protected resource, the block cookies flag MUST also be set.
   - **No redirects:** The user agent MUST NOT follow redirects when fetching this resource.

## Strip URI for Reporting

Before including a URI in a violation report, the user agent applies the following algorithm:

1. If the origin of `uri` is a **globally unique identifier** (for example, `uri` has a scheme of `data`, `blob`, or `filesystem`), then return the ASCII serialization of `uri`'s scheme only.
   - Example: `data:text/javascript,alert(1)` becomes `"data"`
   - Example: `blob:https://example.com/...` becomes `"blob"`

2. If the origin of `uri` is **not the same** as the origin of the protected resource (cross-origin), then return the ASCII serialization of `uri`'s origin only (scheme + host + port).
   - Example: `https://evil.example.com/path/to/script.js` becomes `"https://evil.example.com"` (path stripped)

3. Otherwise (same-origin), return `uri` with any fragment component removed.
   - Example: `https://example.com/page.html#section` becomes `"https://example.com/page.html"` (full path preserved)

## Blocked URI Special Cases

The `blocked-uri` field has several special behaviors:

- **Empty string for inline violations:** When the resource has no URL (inline script, inline style), `blocked-uri` is the empty string `""`.

- **Pre-redirect URL only:** `blocked-uri` will not contain the final location of a resource that was blocked after one or more redirects. It instead will contain only the location that the protected resource requested, before any redirects were followed.

- **Cross-origin truncation:** Per the Strip URI algorithm, cross-origin `blocked-uri` values are truncated to origin only (scheme + host + port). This is why some report `blocked-uri` values lack paths.

- **frame-ancestors special handling:** When generating a violation report for a `frame-ancestors` violation, the user agent MUST NOT include the value of the embedding ancestor as a `blocked-uri` value unless it is same-origin with the protected resource, as disclosing the value of cross-origin ancestors is a violation of the Same-Origin Policy.

- **data/blob/filesystem URIs:** Per the Strip URI algorithm, these are reduced to just the scheme name (e.g., `"data"`, `"blob"`).

## Differences from CSP Level 3

CSP Level 2 uses the `report-uri` directive and legacy `csp-report` JSON format. CSP Level 3 introduces significant changes to reporting:

| Aspect | CSP Level 2 | CSP Level 3 |
|---|---|---|
| **Directive** | `report-uri` | `report-to` (uses Reporting API) |
| **JSON wrapper** | `{"csp-report": {...}}` | `{"type": "csp-violation", "body": {...}, ...}` |
| **Content-Type** | `application/csp-report` | `application/reports+json` |
| **Field naming** | Hyphenated: `blocked-uri`, `document-uri` | camelCase: `blockedURL`, `documentURL` |
| **`disposition` field** | Not present | `"enforce"` or `"report"` |
| **`script-sample` / `sample` field** | Not present | First 40 chars of inline source (requires `'report-sample'`) |
| **`report-to` directive** | Not present | Groups reports via Reporting API endpoint |
| **Deprecation** | Current stable in older browsers | `report-uri` deprecated in favor of `report-to` |

The CSP2 spec itself states: "Development of CSP Level 2 concluded in 2014. Implementors of user-agents are strongly encouraged to base their work on Content Security Policy Level 3."

## Relevance to Vigil

- Vigil's `HandleReport` endpoint accepts `application/csp-report` Content-Type (alongside `application/reports+json` and `application/json`), directly matching the CSP2 spec requirement.
- The `{"csp-report": {...}}` JSON wrapper is the legacy format Vigil detects via `gjson.Get(raw, "csp-report").Exists()` in `parseLegacy()`.
- All required field definitions above (`document-uri`, `effective-directive`, `blocked-uri`) map exactly to the gjson paths Vigil uses: `csp-report.effective-directive`, `csp-report.blocked-uri`, `csp-report.document-uri`.
- Cross-origin URI truncation (Strip URI algorithm) explains why some `blocked-uri` values received by Vigil lack paths — this is spec-mandated browser behavior, not data loss.
- The legacy format lacks a `user_agent` field (unlike the modern Reporting API format), which is why Vigil's `parseLegacy()` always records `"Unknown"` for the browser family.
- Vigil stores the raw JSON body without parsing into model structs, preserving all fields (including optional `source-file`, `line-number`, `column-number`) regardless of which ones are present.

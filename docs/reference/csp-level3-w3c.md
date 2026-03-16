# CSP Level 3 — W3C Working Draft

> Source: https://www.w3.org/TR/CSP3/
> Editor's Draft: https://w3c.github.io/webappsec-csp/
> Spec Status: Working Draft (March 11, 2026)
> Last Fetched: 2026-03-16

## Overview

Content Security Policy Level 3 provides a mechanism by which web developers can control the resources which a particular page can fetch or execute, as well as a number of security-relevant policy decisions. It is a rewrite of CSP Level 2 in terms of the Fetch specification, with enhanced modularity and a stable core for extensibility.

## Directives

### Fetch Directives

These directives control which resources a document may load. Each falls back to `default-src` unless a more specific directive is set.

| Directive | Description | Fallback Chain |
|-----------|-------------|----------------|
| `default-src` | Base policy for all fetch types | (none — this IS the base) |
| `script-src` | Controls external scripts | `default-src` |
| `script-src-elem` | Controls `<script>` elements specifically | `script-src` -> `default-src` |
| `script-src-attr` | Controls inline event handler attributes (e.g., `onclick`) | `script-src` -> `default-src` |
| `style-src` | Controls external stylesheets | `default-src` |
| `style-src-elem` | Controls `<style>` elements and `<link rel=stylesheet>` | `style-src` -> `default-src` |
| `style-src-attr` | Controls inline `style` attributes | `style-src` -> `default-src` |
| `img-src` | Controls image resources | `default-src` |
| `font-src` | Controls font file loading | `default-src` |
| `connect-src` | Controls Fetch, XHR, WebSocket, EventSource connections | `default-src` |
| `media-src` | Controls audio and video resources | `default-src` |
| `object-src` | Controls plugins (`<object>`, `<embed>`) | `default-src` |
| `frame-src` | Controls `<iframe>` element sources | `child-src` -> `default-src` |
| `child-src` | Controls frames and workers (partially deprecated) | `default-src` |
| `worker-src` | Controls Web Worker, Shared Worker, Service Worker sources | `child-src` -> `script-src` -> `default-src` |
| `manifest-src` | Controls web app manifest loading | `default-src` |

Note: `prefetch-src` was previously proposed but is not present in the current Working Draft.

### Document Directives

| Directive | Description |
|-----------|-------------|
| `base-uri` | Restricts the URLs allowed in a document's `<base>` element `href` attribute, preventing base URL manipulation |
| `sandbox` | Applies iframe-like restrictions to the document itself (disables plugins, scripts, top-level navigation, popups, etc.) |

### Navigation Directives

| Directive | Description |
|-----------|-------------|
| `form-action` | Restricts the URLs to which forms can be submitted (does not restrict navigation itself) |
| `frame-ancestors` | Controls which parent documents may embed the current document via `<iframe>`, `<object>`, `<embed>`, or `<frame>` (complementary to `X-Frame-Options`) |

Note: `navigate-to` is defined elsewhere and controls allowed navigation destinations.

### Reporting Directives

| Directive | Description |
|-----------|-------------|
| `report-uri` | **Deprecated.** Legacy reporting endpoint — accepts a URI to which violation reports are POSTed as JSON |
| `report-to` | Preferred. Specifies a Reporting API group name for violation report delivery |

### Other Directives

| Directive | Description |
|-----------|-------------|
| `webrtc` | Controls WebRTC behavior |

## Source Values

### Keywords

| Value | Description |
|-------|-------------|
| `'none'` | Matches nothing; blocks all resources for the directive |
| `'self'` | Matches the current document's origin, including secure upgrades (HTTP->HTTPS, WS->WSS) |
| `'unsafe-inline'` | Allows inline `<script>`, `<style>`, event handler attributes, and `javascript:` URLs |
| `'unsafe-eval'` | Allows dynamic code execution: `Function()`, `setTimeout(string)`, `setInterval(string)`, and similar |
| `'unsafe-hashes'` | Allows inline event handlers and style attributes that match specified hashes |
| `'wasm-unsafe-eval'` | Allows WebAssembly byte compilation (`WebAssembly.compile()`, etc.) |
| `'strict-dynamic'` | Allows scripts loaded by already-trusted (nonce/hash) scripts; ignores allowlist sources |
| `'report-sample'` | Includes the first 40 characters of the blocked inline content in violation reports |
| `'unsafe-allow-redirects'` | Allows redirects in source matching |
| `'trusted-types-eval'` | Allows Trusted Types policy object creation |
| `'unsafe-webtransport-hashes'` | WebTransport hash support |

### Hash and Nonce Sources

| Format | Description |
|--------|-------------|
| `'nonce-<base64-value>'` | Matches elements with a matching `nonce` attribute; random per page load |
| `'sha256-<base64-value>'` | Matches content whose SHA-256 hash matches the base64-encoded value |
| `'sha384-<base64-value>'` | Matches content whose SHA-384 hash matches |
| `'sha512-<base64-value>'` | Matches content whose SHA-512 hash matches |
| `'report-sha256'` | Enable hash reporting for external resources with matching subresource integrity |
| `'report-sha384'` | (same, SHA-384) |
| `'report-sha512'` | (same, SHA-512) |

### Host and Scheme Sources

| Format | Example | Description |
|--------|---------|-------------|
| Host | `example.com` | Matches the exact host |
| Wildcard host | `*.example.com` | Matches any subdomain of the host |
| Full URI | `https://example.com:443/path` | Matches scheme, host, port, and path prefix |
| Scheme | `https:`, `data:`, `blob:`, `wss:` | Matches any resource loaded via the scheme |

**Matching behavior**: Insecure schemes/ports match their secure variants (e.g., `http://example.com:80` matches both HTTP and HTTPS). Nonce source expressions are restricted to HTTP(S) schemes by default.

## Violation Report Generation (Section 5)

### Violation Object Structure

Each violation internally maintains:

| Property | Type | Description |
|----------|------|-------------|
| `global object` | object | The context whose policy was violated |
| `url` | URL | The global object's URL |
| `status` | non-negative integer | HTTP status code of the resource |
| `resource` | null, string, or URL | One of: `null`, `"inline"`, `"eval"`, `"wasm-eval"`, `"trusted-types-policy"`, `"trusted-types-sink"`, or a URL |
| `referrer` | null or URL | Referring document URL |
| `policy` | object | The violated policy object |
| `disposition` | string | Either `"enforce"` or `"report"` |
| `effective directive` | string | The directive name that caused the violation |
| `source file` | null or URL | URL of the script that triggered the violation |
| `line number` | non-negative integer | Line in the source file |
| `column number` | non-negative integer | Column in the source file |
| `element` | null or Element | The DOM element that triggered the violation |
| `sample` | string | Empty unless populated; first 40 characters of inline content when `'report-sample'` is present |

### CSPViolationReportBody WebIDL

The report body sent via the Reporting API (`report-to`) contains:

| Property | Type | Description |
|----------|------|-------------|
| `documentURL` | string | The stripped URL of the document where the violation occurred |
| `referrer` | string | The referrer of the document |
| `blockedURL` | string | The blocked resource identifier (see Obtain Blocked URI) |
| `effectiveDirective` | string | The directive whose enforcement caused the violation |
| `originalPolicy` | string | The full original policy string |
| `sourceFile` | string | URL of the script that triggered the violation |
| `sample` | string | First 40 characters of the violating inline content (if `'report-sample'` is present) |
| `disposition` | string | `"enforce"` or `"report"` |
| `statusCode` | unsigned short | HTTP status code of the document |
| `lineNumber` | unsigned long | Line number in the source file |
| `columnNumber` | unsigned long | Column number in the source file |

### Report Generation Algorithm

The algorithm for creating a violation from a global object, policy, and directive:

1. Create a new violation object with `resource` set to `null`.
2. Extract the source file URL, line number, and column number from the currently executing script context (if available).
3. Set the `referrer` from the Window's document referrer (if the global is a Window).
4. Populate the HTTP `status` code from the associated resource.
5. Return the populated violation.

For violations from a request and policy:

1. Determine the effective directive via request destination matching.
2. Invoke the global/policy/directive creation algorithm above.
3. Set `resource` to the request's URL (pre-redirect URL, not the final URL).
4. Return the violation.

### Obtain Blocked URI Algorithm (Section 5.2)

The blocked resource identifier is derived from the violation's `resource` property:

| Resource Value | blockedURL Value | When Used |
|----------------|------------------|-----------|
| `null` | `""` (empty string) | No specific resource identified |
| `"inline"` | `"inline"` | Inline `<script>`, `<style>`, event handlers, `javascript:` URLs |
| `"eval"` | `"eval"` | Dynamic code execution attempts |
| `"wasm-eval"` | `"wasm-eval"` | WebAssembly byte compilation |
| `"trusted-types-policy"` | `"trusted-types-policy"` | Trusted Types policy creation violations |
| `"trusted-types-sink"` | `"trusted-types-sink"` | Trusted Types sink violations |
| A URL | The URL (stripped for reporting) | External resources blocked by fetch directives |

### Deprecated Serialization (Section 5.3)

The legacy `csp-report` JSON format used with `report-uri`. The report is POSTed with `Content-Type: application/csp-report`.

**Field mapping from modern to legacy format:**

| Modern (CSPViolationReportBody) | Legacy (csp-report) | Notes |
|---------------------------------|---------------------|-------|
| `documentURL` | `document-uri` | Stripped for cross-origin privacy |
| `referrer` | `referrer` | Same value |
| `blockedURL` | `blocked-uri` | Same value |
| `effectiveDirective` | `violated-directive` | **Renamed** in legacy format |
| `originalPolicy` | `original-policy` | Full policy string |
| `sourceFile` | `source-file` | Script URL |
| `sample` | (not in legacy) | Only available in modern format |
| `disposition` | `disposition` | `"enforce"` or `"report"` |
| `statusCode` | `status-code` | HTTP response code |
| `lineNumber` | `line-number` | Source line |
| `columnNumber` | `column-number` | Source column |

**Legacy report JSON structure:**

```json
{
  "csp-report": {
    "document-uri": "https://example.com/page",
    "referrer": "https://example.com/",
    "violated-directive": "script-src",
    "original-policy": "script-src 'self'; report-uri /csp-report",
    "blocked-uri": "https://evil.com/script.js",
    "status-code": 200,
    "source-file": "https://example.com/page",
    "line-number": 42,
    "column-number": 10,
    "disposition": "enforce"
  }
}
```

### Strip URL for Reporting Algorithm (Section 5.4)

URLs included in violation reports are stripped to prevent unintended leakage of sensitive information:

1. Accept a URL for processing.
2. Produce an origin-only representation (scheme + host + port), removing path, query string, and fragment.
3. This prevents cross-origin path and query parameter information from being leaked in violation reports sent to third-party endpoints.

This algorithm is applied to `document-uri` / `documentURL`, `source-file` / `sourceFile`, and external `blocked-uri` / `blockedURL` values.

### Report a Violation Algorithm (Section 5.5)

The steps for dispatching a violation report:

1. Fire a `securitypolicyviolation` event at the violation's global object (DOM event, allowing JavaScript monitoring).
2. Determine applicable reporting endpoints:
   - If the policy has a `report-to` directive: use the Reporting API to queue a report to the named group.
   - If the policy has a `report-uri` directive: POST the deprecated serialization (Section 5.3) to the specified URI(s).
3. Construct the violation report body:
   - Set `documentURL` to the stripped document URL.
   - Set `blockedURL` via the Obtain Blocked URI algorithm.
   - Set `effectiveDirective` to the violated directive name.
   - Set `originalPolicy` to the full policy string.
   - Set `disposition` to `"enforce"` or `"report"`.
   - Set `statusCode` to the document's HTTP status code.
   - Populate `sourceFile`, `lineNumber`, `columnNumber` if available.
   - If `'report-sample'` is in the directive's source list, include `sample` (first 40 characters of the inline content).
   - Set `referrer` from the document.
4. Queue the report:
   - For `report-to`: queue via the Reporting API infrastructure.
   - For `report-uri`: POST JSON with `Content-Type: application/csp-report`.

## report-uri vs report-to Interaction

- `report-to` is the preferred mechanism and takes priority when both are present.
- `report-uri` is deprecated but remains widely supported as a fallback.
- When a policy contains both `report-to` and `report-uri`:
  - User agents that support the Reporting API use `report-to`.
  - User agents that do not support the Reporting API fall back to `report-uri`.
- The spec recommends including both directives during the transition period for maximum browser coverage.
- `report-to` uses the Reporting API (`Report-To` HTTP header defines endpoint groups); `report-uri` uses direct POST to a URL.

## Blocking Algorithms Summary

### Request Pre-blocking

"Should request be blocked by Content Security Policy?":

1. Iterate through enforce-disposition policies only.
2. Check if the request violates any policy via the matching algorithm.
3. Report violations.
4. Return `Blocked` or `Allowed`.

### Response Post-blocking

"Should response to request be blocked by Content Security Policy?":

1. Iterate all policies and their directives.
2. Run each directive's post-request check.
3. Report violations; only enforce-disposition policies cause actual blocking.
4. Return `Blocked` or `Allowed`.

### Inline Element Blocking

"Should element's inline type behavior be blocked by Content Security Policy?":

1. Accept element, type (`script`, `style`, event handler, etc.), and source string.
2. Check all policies' directives via the inline check algorithm.
3. Create violation with `"inline"` as the resource.
4. Extract first 40 characters as sample if `'report-sample'` is present in the directive.
5. Return `Blocked` or `Allowed`.

## Key Architectural Notes

- **Multiple Policies**: A single resource can have multiple CSP policies (via multiple headers and/or `<meta>` elements). ALL policies must be satisfied for a resource to load.
- **Path Matching**: CSP3 supports path-based restrictions; paths are matched per RFC 3986 semantics.
- **`frame-src` undeprecated**: Was deprecated in CSP2, restored in CSP3 alongside `child-src`.
- **Secure upgrade matching**: `http://example.com:80` in a policy now matches `https://example.com:443`.

## Relevance to Vigil

- Vigil receives reports at both legacy (`report-uri`) and modern (`report-to`) endpoints.
- **CSPViolationReportBody** defines all fields Vigil may encounter in modern reports (via the Reporting API).
- **Deprecated serialization (Section 5.3)** defines the legacy `csp-report` JSON format Vigil handles — note the field name differences (`effectiveDirective` vs `violated-directive`, `documentURL` vs `document-uri`, etc.).
- The **strip URL algorithm** explains why some `blocked-uri` / `blockedURL` values are truncated to origin-only, losing path information.
- **Special `blockedURL` values** (`inline`, `eval`, `wasm-eval`, `trusted-types-policy`, `trusted-types-sink`) appear in Vigil's aggregated Slack reports and are not URLs — they are sentinel strings indicating the type of violation.
- The **sample** field (first 40 characters of violating inline content) is only populated when `'report-sample'` is in the policy's source list.
- **Disposition** distinguishes enforced violations (`"enforce"`) from report-only violations (`"report"`), corresponding to the `Content-Security-Policy` vs `Content-Security-Policy-Report-Only` headers.
- **Multiple policies** means Vigil may receive multiple reports for a single page load — one per violated policy.

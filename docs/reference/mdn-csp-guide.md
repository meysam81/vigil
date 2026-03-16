# Content Security Policy (CSP) — MDN Guide

> Source: https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CSP
> Spec Status: Living Standard
> Last Fetched: 2026-03-16

## Overview

Content Security Policy (CSP) is a feature that helps prevent or minimize the risk of certain types of security threats. It consists of a series of instructions from a website to a browser, which instruct the browser to place restrictions on the things that the code comprising the site is allowed to do. The primary use case is controlling which resources — in particular JavaScript — a document is allowed to load, serving as a defense against cross-site scripting (XSS) attacks in which an attacker injects malicious code into the victim's site.

A CSP can also defend against clickjacking (via `frame-ancestors`) and help ensure pages are loaded over HTTPS (via `upgrade-insecure-requests`).

## Directives by Category

### Fetch Directives

Fetch directives specify categories of resources a document is allowed to load. Each falls back to `default-src` if not explicitly listed.

| Directive | Description |
|-----------|-------------|
| `default-src` | Fallback for all other fetch directives. |
| `child-src` | Valid sources for web workers and nested browsing contexts (`<frame>`, `<iframe>`). Fallback for `frame-src` and `worker-src`. |
| `connect-src` | Restricts URLs which can be loaded using script interfaces (fetch, XHR, WebSocket, EventSource, etc.). |
| `fenced-frame-src` | Valid sources for nested browsing contexts loaded into `<fencedframe>` elements. *(Experimental)* |
| `font-src` | Valid sources for fonts loaded using `@font-face`. |
| `frame-src` | Valid sources for nested browsing contexts in `<frame>` and `<iframe>`. |
| `img-src` | Valid sources of images and favicons. |
| `manifest-src` | Valid sources of application manifest files. |
| `media-src` | Valid sources for loading `<audio>`, `<video>`, and `<track>` elements. |
| `object-src` | Valid sources for `<object>` and `<embed>` elements. |
| `prefetch-src` | Valid sources to be prefetched or prerendered. *(Deprecated, Non-standard)* |
| `script-src` | Valid sources for JavaScript and WebAssembly. Fallback for `script-src-elem` and `script-src-attr`. |
| `script-src-elem` | Valid sources for JavaScript `<script>` elements. |
| `script-src-attr` | Valid sources for JavaScript inline event handlers. |
| `style-src` | Valid sources for stylesheets. Fallback for `style-src-elem` and `style-src-attr`. |
| `style-src-elem` | Valid sources for `<style>` elements and `<link rel="stylesheet">` elements. |
| `style-src-attr` | Valid sources for inline styles applied to individual DOM elements. |
| `worker-src` | Valid sources for `Worker`, `SharedWorker`, or `ServiceWorker` scripts. |

**Fallback hierarchy:**

- `default-src` is the fallback for all other fetch directives
- `script-src` is the fallback for `script-src-elem` and `script-src-attr`
- `style-src` is the fallback for `style-src-elem` and `style-src-attr`
- `child-src` is the fallback for `frame-src` and `worker-src`

### Document Directives

| Directive | Description |
|-----------|-------------|
| `base-uri` | Restricts the URLs which can be used in a document's `<base>` element. |
| `sandbox` | Enables a sandbox for the requested resource, similar to the `<iframe>` sandbox attribute. |

### Navigation Directives

| Directive | Description |
|-----------|-------------|
| `form-action` | Restricts the URLs which can be used as the target of form submissions from a given context. |
| `frame-ancestors` | Specifies valid parents that may embed a page using `<frame>`, `<iframe>`, `<object>`, or `<embed>`. Effective replacement for `X-Frame-Options`. |

Note: `navigate-to` was proposed but is not listed in the current MDN reference as a supported directive.

### Reporting Directives

| Directive | Description |
|-----------|-------------|
| `report-to` | Provides the browser a token identifying a reporting endpoint (defined via `Reporting-Endpoints` header) to send CSP violation reports to. Uses the modern Reporting API. |
| `report-uri` | Provides a URL where CSP violation reports should be sent. *(Deprecated)* Sends a slightly different JSON format with `Content-Type: application/csp-report`. |

When using `report-to`, MDN recommends also specifying `report-uri` for broader browser support:

```
Content-Security-Policy: ...; report-uri https://endpoint.example.com; report-to endpoint_name
```

### Other Directives

| Directive | Description |
|-----------|-------------|
| `upgrade-insecure-requests` | Instructs browsers to treat all insecure URLs (HTTP) as though replaced with HTTPS. |
| `require-trusted-types-for` | Enforces Trusted Types at DOM XSS injection sinks. |
| `trusted-types` | Specifies an allowlist of Trusted Types policies, locking down DOM XSS injection sinks. |
| `block-all-mixed-content` | Prevents loading any assets using HTTP when the page is loaded using HTTPS. *(Deprecated)* |

## Source Values

| Source Value | Applicable To | Description |
|-------------|---------------|-------------|
| `'none'` | All fetch directives | Completely blocks the specific resource type. Cannot be combined with other source expressions (they are ignored if present). |
| `'self'` | All fetch directives | Allows resources from the same origin as the document. Secure upgrades allowed (HTTP to HTTPS, WS to WSS). |
| `'unsafe-inline'` | `script-src`, `style-src` | Allows inline JavaScript and CSS. Defeats much of CSP's purpose. Ignored when nonce or hash expressions are present. |
| `'unsafe-eval'` | `script-src` | Allows dynamic code execution via `Function()` constructor, `setTimeout(string)`, `setInterval(string)`, and similar APIs. Unlike `'unsafe-inline'`, this still works even when nonce or hash expressions are present. |
| `'unsafe-hashes'` | `script-src`, `style-src` | Allows inline event handlers and style attributes using hash expressions. Safer than `'unsafe-inline'` but still considered unsafe. |
| `'wasm-unsafe-eval'` | `script-src` | Allows WebAssembly compilation via `WebAssembly.compileStreaming()`. Safer alternative to `'unsafe-eval'` for Wasm use cases. |
| `'strict-dynamic'` | `script-src` | Trust from nonce/hash extends to dynamically loaded scripts (scripts loaded by trusted scripts). When present, host-source, scheme-source, `'self'`, and `'unsafe-inline'` are ignored. |
| `'report-sample'` | `script-src`, `style-src` | When violations occur, the report includes the first 40 characters of the blocked resource in a `sample` property. |
| `'inline-speculation-rules'` | `script-src` | Allows inline `<script type="speculationrules">` elements. |
| `'trusted-types-eval'` | `script-src` | Enables dynamic code execution functions only when Trusted Types are enforced and passed instead of strings. |
| `nonce-<base64>` | `<script>`, `<style>` | Server generates a random value per HTTP response. Element must have matching `nonce` attribute. When present, `'unsafe-inline'` is ignored. |
| `sha256-<hash>` | `<script>`, `<style>` | Browser hashes element contents; loads only if hash matches. External scripts must also include the `integrity` attribute. |
| `sha384-<hash>` | `<script>`, `<style>` | Same as sha256 but using SHA-384 algorithm. |
| `sha512-<hash>` | `<script>`, `<style>` | Same as sha256 but using SHA-512 algorithm. |
| Host source | All fetch directives | Format: `<scheme>://<host>:<port>/<path>`. Scheme, port, and path are optional. Wildcards supported: `*.example.com` (subdomains), `example.com:*` (all ports). Paths ending in `/` match prefixes. |
| Scheme source | All fetch directives | Format: `<scheme>:` (colon required). Examples: `https:`, `http:`, `data:`, `blob:`, `mediastream:`, `filesystem:`. Secure upgrades allowed (`http:` permits HTTPS, `ws:` permits WSS). |

## Report-Only Mode

CSP can be deployed in report-only mode using the `Content-Security-Policy-Report-Only` header. The policy is not enforced, but violations are sent to the reporting endpoint. This is useful for testing policies before enforcement.

Key behaviors:

- If both `Content-Security-Policy` and `Content-Security-Policy-Report-Only` are present, both policies are honored: the `Content-Security-Policy` policy is enforced, and the `Content-Security-Policy-Report-Only` policy generates reports but is not enforced.
- A report-only policy cannot be delivered via a `<meta>` element.
- Violation reports from report-only mode have `disposition: "report"` (vs `disposition: "enforce"` for enforced policies).
- The `sandbox` directive is ignored when delivered in a report-only policy.

## Meta Tag Limitations

CSP can be specified using the `<meta http-equiv="Content-Security-Policy">` tag, useful for single-page apps with only static resources. However, the following limitations apply:

- `frame-ancestors` is **not supported** via meta tag.
- `report-uri` is **not supported** via meta tag.
- `report-to` is **not supported** via meta tag.
- `sandbox` is **not supported** via meta tag.
- Source values incompatible with meta delivery are silently ignored (browser error: "Ignoring source '%1$S' (Not supported when delivered via meta element)").
- Report-only mode (`Content-Security-Policy-Report-Only`) cannot be delivered via meta tag.

## Violation Report Lifecycle

1. **Policy delivery**: The server sends a CSP header (or report-only header) with the HTTP response. The `report-to` directive names a reporting endpoint defined in the `Reporting-Endpoints` header.

2. **Violation detection**: The browser evaluates every resource load, inline script execution, dynamic code execution call, etc. against the active CSP policy. If a resource or action does not match any allowed source expression in the applicable directive, a violation occurs.

3. **Report generation**: The browser constructs a JSON report object. For the modern Reporting API (`report-to`), the report structure is:

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
     "user_agent": "Mozilla/5.0 ..."
   }
   ```

4. **Report delivery**: The browser sends the report via HTTP `POST` to the designated endpoint. Modern format uses `Content-Type: application/reports+json`. Legacy `report-uri` format uses `Content-Type: application/csp-report`.

5. **Batching**: The Reporting API may batch multiple reports into a single POST request as a JSON array.

## Blocked URI Edge Cases

The `blockedURL` field in violation reports has several special values depending on what triggered the violation:

| Blocked URI Value | Trigger |
|-------------------|---------|
| `inline` | Inline `<script>` or `<style>` elements blocked by missing `'unsafe-inline'`, nonce, or hash in `script-src` / `style-src`. Also applies to inline event handlers (e.g., `onclick`). |
| `eval` | Calls to dynamic code execution APIs (`Function()`, `setTimeout(string)`, `setInterval(string)`) blocked by missing `'unsafe-eval'` in `script-src`. |
| `data` | Resources loaded via `data:` URIs when `data:` is not listed as an allowed scheme source. |
| `blob` | Resources loaded via `blob:` URIs when `blob:` is not listed as an allowed scheme source. |
| `wasm-eval` | WebAssembly compilation blocked by missing `'wasm-unsafe-eval'` in `script-src`. |
| Actual URL | When an external resource at a specific URL is blocked, the full or partial URL appears. |

Note: The `effectiveDirective` field in reports shows the specific sub-directive that was violated (e.g., `script-src-elem` rather than `script-src`), even if only the parent directive was explicitly set in the policy.

## Nonce and Hash Behavior

### Nonces

- A nonce is the recommended approach for restricting `<script>` and `<style>` loading.
- The server generates a random value for every HTTP response and includes it in the `script-src` and/or `style-src` directive: `script-src 'nonce-<value>'`.
- The same nonce value is set as the `nonce` attribute on all `<script>` / `<style>` tags the server intends to allow.
- The browser compares the two values and loads the resource only if they match.
- The nonce **must be different for every HTTP response** and must not be predictable.
- Works with both external and inline scripts.
- When a nonce is present in the directive, `'unsafe-inline'` is **ignored** by browsers.

### Hashes

- The server calculates a hash of the script/style contents using SHA-256, SHA-384, or SHA-512.
- The Base64-encoded hash is added to the directive: `script-src 'sha256-<base64>'`.
- The browser hashes the element contents and loads only if the hash matches.
- External scripts **must also include the `integrity` attribute** with the same hash value for this to work.
- Each script/style element needs its own separate hash.
- Hashes are better suited for static pages or client-side rendering since both CSP and content can be static.
- When a hash is present in the directive, `'unsafe-inline'` is **ignored** by browsers.

### Interaction with `'strict-dynamic'`

- `'strict-dynamic'` extends the trust granted by a nonce or hash to scripts that are dynamically loaded by the trusted script.
- When `'strict-dynamic'` is present, host-source expressions, scheme-source expressions, `'self'`, and `'unsafe-inline'` are all **ignored**.
- Without `'strict-dynamic'`, a nonced/hashed script that dynamically creates and appends a new `<script>` element will have that new script **blocked** (it has no nonce/hash of its own).
- With `'strict-dynamic'`, the dynamically loaded script is allowed because it was loaded by a trusted (nonced/hashed) parent script.
- This trust propagation is transitive: scripts loaded by dynamically loaded scripts are also trusted.
- Warning from MDN: `'strict-dynamic'` within a directive with no valid nonce or hash might block all scripts from loading.

## Relevance to Vigil

- This guide provides context for understanding what triggers CSP violation reports that Vigil collects.
- Vigil receives reports from all directive categories when violations occur — fetch, document, navigation, and reporting directives can all produce violations.
- Report-Only mode reports have `disposition: "report"` (not `"enforce"`), allowing Vigil to distinguish between enforced blocks and monitoring-only observations.
- Understanding source values helps interpret the `originalPolicy` field in reports: keywords like `'self'`, `'unsafe-inline'`, `'strict-dynamic'`, nonces, and hashes all appear in this field.
- Blocked URI edge cases explain the special values Vigil sees in `blockedURL` — `"inline"` for inline scripts/styles, `"eval"` for dynamic code execution calls, and `"data"` / `"blob"` for those URI schemes.
- The `effectiveDirective` field in reports reflects the specific sub-directive (e.g., `script-src-elem`) even when only the parent directive (`script-src` or `default-src`) was set in the policy.
- The `sample` property (first 40 characters of blocked content) is only populated when `'report-sample'` is included in the directive.
- Legacy `report-uri` reports arrive with `Content-Type: application/csp-report` and a different JSON structure than modern `report-to` reports (`Content-Type: application/reports+json`). Vigil must handle both formats.

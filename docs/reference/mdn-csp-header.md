# Content-Security-Policy Header -- MDN Reference

> Source: <https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy>
> Spec Status: Living Standard (Content Security Policy Level 3)
> Last Fetched: 2026-03-16

## Overview

The HTTP `Content-Security-Policy` response header allows website administrators to control which resources the user agent is allowed to load for a given page. With a few exceptions, policies mostly involve specifying server origins and script endpoints. This helps guard against cross-site scripting (XSS) attacks.

Header Type: Response header
Specification: [Content Security Policy Level 3](https://w3c.github.io/webappsec-csp/#csp-header)
Browser Baseline: Widely available since August 2016.

## Syntax

```http
Content-Security-Policy: <policy-directive>; <policy-directive>
```

Where `<policy-directive>` consists of `<directive> <value>` with no internal punctuation. Multiple directives are separated by semicolons.

## Complete Directive List

### Fetch Directives

Fetch directives control the locations from which certain resource types may be loaded.

| Directive | Description |
|-----------|-------------|
| `default-src` | Serves as a fallback for all other fetch directives. If a specific fetch directive is absent, the browser uses `default-src` instead. |
| `script-src` | Specifies valid sources for JavaScript and WebAssembly resources. Falls back to `default-src`. Acts as fallback for `script-src-elem` and `script-src-attr`. |
| `script-src-elem` | Specifies valid sources for JavaScript `<script>` elements. Falls back to `script-src`. |
| `script-src-attr` | Specifies valid sources for JavaScript inline event handlers. Falls back to `script-src`. |
| `style-src` | Specifies valid sources for stylesheets. Falls back to `default-src`. Acts as fallback for `style-src-elem` and `style-src-attr`. |
| `style-src-elem` | Specifies valid sources for stylesheets `<style>` elements and `<link>` elements with `rel="stylesheet"`. Falls back to `style-src`. |
| `style-src-attr` | Specifies valid sources for inline styles applied to individual DOM elements. Falls back to `style-src`. |
| `img-src` | Specifies valid sources of images and favicons. Falls back to `default-src`. |
| `font-src` | Specifies valid sources for fonts loaded using `@font-face`. Falls back to `default-src`. |
| `connect-src` | Restricts URLs which can be loaded using script interfaces (fetch, XHR, WebSocket, EventSource, etc.). Falls back to `default-src`. |
| `media-src` | Specifies valid sources for loading media using `<audio>`, `<video>`, and `<track>` elements. Falls back to `default-src`. |
| `object-src` | Specifies valid sources for `<object>` and `<embed>` elements. Falls back to `default-src`. |
| `frame-src` | Specifies valid sources for nested browsing contexts loaded into `<frame>` and `<iframe>` elements. Falls back to `child-src`. |
| `child-src` | Defines valid sources for web workers and nested browsing contexts loaded using `<frame>` and `<iframe>`. Falls back to `default-src`. Acts as fallback for `frame-src` and `worker-src`. |
| `worker-src` | Specifies valid sources for `Worker`, `SharedWorker`, or `ServiceWorker` scripts. Falls back to `child-src`. |
| `manifest-src` | Specifies valid sources of application manifest files. Falls back to `default-src`. |
| `prefetch-src` | Specifies valid sources to be prefetched or prerendered. **Deprecated, Non-standard.** |
| `fenced-frame-src` | Specifies valid sources for nested browsing contexts loaded into `<fencedframe>` elements. **Experimental.** |

#### Fallback Chain

```
default-src
  +-- script-src
  |     +-- script-src-elem
  |     +-- script-src-attr
  +-- style-src
  |     +-- style-src-elem
  |     +-- style-src-attr
  +-- img-src
  +-- font-src
  +-- connect-src
  +-- media-src
  +-- object-src
  +-- manifest-src
  +-- child-src
        +-- frame-src
        +-- worker-src
```

### Document Directives

Govern the properties of a document or worker environment to which a policy applies.

| Directive | Description |
|-----------|-------------|
| `base-uri` | Restricts URLs which can be used in a document's `<base>` element. |
| `sandbox` | Enables a sandbox for the requested resource similar to the `<iframe>` `sandbox` attribute. |

### Navigation Directives

Govern to which locations a user can navigate or submit a form.

| Directive | Description |
|-----------|-------------|
| `form-action` | Restricts URLs which can be used as the target of form submissions from a given context. |
| `frame-ancestors` | Specifies valid parents that may embed a page using `<frame>`, `<iframe>`, `<object>`, or `<embed>`. Setting to `'none'` is similar to `X-Frame-Options: DENY`. |

Note: `navigate-to` was proposed in CSP Level 3 but is not mentioned on the current MDN page as a supported directive.

### Reporting Directives

Control the destination URL for CSP violation reports.

| Directive | Description |
|-----------|-------------|
| `report-to` | Provides the browser with a token identifying the reporting endpoint or group of endpoints to send CSP violation information to. Endpoints are defined via the `Reporting-Endpoints` or deprecated `Report-To` HTTP response headers. Intended to replace `report-uri`. |
| `report-uri` | Provides the browser with a URL where CSP violation reports should be sent. **Deprecated** -- superseded by `report-to`. |

### Other Directives

| Directive | Description |
|-----------|-------------|
| `upgrade-insecure-requests` | Instructs user agents to treat all of a site's insecure URLs (HTTP) as though they have been replaced with secure URLs (HTTPS). Intended for websites with large numbers of insecure legacy URLs. |
| `require-trusted-types-for` | Enforces Trusted Types at the DOM XSS injection sinks. |
| `trusted-types` | Used to specify an allowlist of Trusted Types policies. Allows applications to lock down DOM XSS injection sinks to only accept non-spoofable, typed values in place of strings. |

### Deprecated Directives

| Directive | Description |
|-----------|-------------|
| `block-all-mixed-content` | Prevents loading any assets using HTTP when the page is loaded using HTTPS. **Deprecated.** |
| `report-uri` | See Reporting Directives above. **Deprecated.** |

## Source List Values

All fetch directives may be specified as the single value `'none'` or one or more source expression values.

**Quoting rule:** `<host-source>` and `<scheme-source>` must be unquoted. All other values must be enclosed in single quotes.

### `'none'`

Blocks all resources of the specified type. No resources are allowed to load.

### `'self'`

Resources of the given type may only be loaded from the same origin as the document. Secure upgrades are allowed: if the document is served from `http://example.com`, CSP `'self'` also permits resources from `https://example.com`.

### `'unsafe-inline'`

Allows inline JavaScript and CSS. For scripts, this unblocks inline `<script>` tags, inline event handler attributes, and `javascript:` URLs. For styles, this unblocks inline `<style>` tags and `style` attributes. **Warning:** developers should avoid this as it defeats much of the purpose of CSP.

### `'unsafe-eval'`

Allows dynamic evaluation of strings as JavaScript, including the `Function()` constructor and timer functions with string arguments. By default, CSP disables these functions. **Warning:** developers should avoid this as it defeats much of the purpose of CSP.

### `'unsafe-hashes'`

Allows the browser to use hash expressions to authorize specific inline event handlers and style attributes. Example: `script-src 'unsafe-hashes' 'sha256-cd9827ad...'`. If the hash matches the content of an inline event handler attribute value or `style` attribute value, the code is allowed to execute. **Warning:** this is unsafe because it enables an attack where inline event handler content can be injected as an inline `<script>` element, but it is much safer than `'unsafe-inline'`.

### `'wasm-unsafe-eval'`

Allows WebAssembly compilation (e.g., `WebAssembly.compileStreaming()`). By default, CSP disables WebAssembly compilation functions. This is a much safer alternative to `'unsafe-eval'` since it does not enable general JavaScript evaluation.

### `'strict-dynamic'`

Makes trust conferred on a script by a nonce or hash extend to scripts that are dynamically loaded by that script (e.g., via `Document.createElement()` and `Node.appendChild()`). When present, the following source expressions are all ignored: `<host-source>`, `<scheme-source>`, `'self'`, and `'unsafe-inline'`.

### `'report-sample'`

If included in a directive controlling scripts or styles, when the directive causes the browser to block inline scripts, inline styles, or event handler attributes, the violation report will contain a `sample` property with the first 40 characters of the blocked resource.

### `'inline-speculation-rules'`

Allows inline `<script>` elements for speculation rules (i.e., `<script>` elements with `type="speculationrules"`). By default, if CSP contains `default-src` or `script-src`, inline JavaScript is not allowed; this keyword allows the browser to load such elements.

### `'trusted-types-eval'`

Allows evaluation when Trusted Types are enforced. Re-enables dynamic code evaluation functions, but only when Trusted Types are passed instead of strings. The transformation function has the opportunity to sanitize input. Must be used instead of `'unsafe-eval'` when using these methods with trusted types.

### `nonce-<base64-value>`

A cryptographic nonce (number used once) that the server generates for every HTTP response. Format: `'nonce-<value>'` where `<value>` is a Base64 or URL-safe Base64 string. The server includes the same nonce as the `nonce` attribute of allowed `<script>` or `<style>` elements. The browser compares the CSP nonce against the element's attribute value; the resource loads only if they match. If a directive contains a nonce and `'unsafe-inline'`, the browser ignores `'unsafe-inline'`.

Example: `'nonce-416d1177-4d12-4e3b-b7c9-f6c409789fb8'`

### `<hash-algorithm>-<base64-value>`

A hash of the inline script or style content. Format: `'<algorithm>-<hash>'`. Supported algorithms: `sha256`, `sha384`, `sha512`. The browser hashes the contents of `<script>` and `<style>` elements and compares against hashes in the CSP directive; the resource loads only if there is a match. For external resources loaded via `src`, the element must also have the `integrity` attribute set. If a directive contains a hash and `'unsafe-inline'`, the browser ignores `'unsafe-inline'`.

Example: `'sha256-cd9827ad...'`

### `<host-source>`

URL or IP address of a host that is a valid source. **Unquoted.** Scheme, port number, and path are optional.

- If scheme is omitted, the document's origin scheme is used.
- Secure upgrades allowed (e.g., `http://example.com` also permits `https://example.com`).
- Wildcards (`*`) allowed for subdomains, host address, and port number.
  - `http://*.example.com` permits resources from any subdomain over HTTP or HTTPS.
- Paths ending in `/` match any path they are a prefix of.
  - `example.com/api/` permits `example.com/api/users/new`.
- Paths not ending in `/` are matched exactly.

### `<scheme-source>`

A scheme such as `https:`. **Unquoted.** The colon is required.

Common schemes: `https:`, `http:`, `data:`, `blob:`, `ws:`, `wss:`.

Secure upgrades allowed: `http:` also permits resources loaded using HTTPS.

## Special Keywords -- Detailed Behavior

### `'strict-dynamic'`

When a policy includes `'strict-dynamic'`, trust is propagated from a nonced or hashed script to any scripts it dynamically loads. This means you only need to trust the initial script; all scripts it creates via DOM APIs inherit trust automatically. Critically, when `'strict-dynamic'` is present, host-based and scheme-based allowlists are ignored, as are `'self'` and `'unsafe-inline'`. This makes it a powerful mechanism for migration: you can add `'strict-dynamic'` alongside existing allowlists for backward compatibility with older browsers that don't understand it, while newer browsers enforce the stricter nonce/hash-based policy.

### `'report-sample'`

When `'report-sample'` is present in a CSP directive (e.g., `script-src 'self' 'report-sample'`), and that directive blocks an inline script, style, or event handler, the resulting violation report will include a `sample` field containing the first 40 characters of the blocked code. Without this keyword, the `sample` field is empty or absent. This is critical for debugging CSP violations because it reveals what code was actually blocked.

## Examples

### 1. HTTPS-only, no inline code

```http
Content-Security-Policy: default-src https:
```

Only allows resource loading over HTTPS. Inline scripts and styles are blocked.

### 2. Allow inline code and HTTPS, disable plugins

```http
Content-Security-Policy: default-src https: 'unsafe-eval' 'unsafe-inline'; object-src 'none'
```

Resources loaded only over HTTPS. Inline code is allowed. Plugins (`<object>`, `<embed>`) are disabled.

### 3. Report-only mode for testing

```http
Reporting-Endpoints: csp-endpoint="https://example.com/csp-reports"
Content-Security-Policy-Report-Only: default-src https:; report-uri /csp-violation-report-url/; report-to csp-endpoint
```

Tests CSP violations without blocking code execution. Both `report-uri` and `report-to` are specified for backward compatibility since `report-to` is not yet broadly supported by all browsers.

## Multiple Policies

The CSP mechanism allows multiple policies for a resource via the `Content-Security-Policy` header, the `Content-Security-Policy-Report-Only` header, or the `<meta>` element. Adding additional policies can only **further restrict** the capabilities of the protected resource; the strictest policy is enforced.

```http
Content-Security-Policy: default-src 'self' http://example.com; connect-src 'none';
Content-Security-Policy: connect-src http://example.com/; script-src http://example.com/
```

Even though the second policy allows `connect-src http://example.com/`, the first policy's `connect-src 'none'` is enforced because the intersection is used.

## CSP in Workers

Workers are generally **not** governed by the CSP of the document that created them. To specify CSP for a worker, set a `Content-Security-Policy` response header for the request that delivers the worker script itself. The exception is if the worker script's origin is a globally unique identifier (e.g., `data:` or `blob:` URL), in which case the worker inherits the CSP of the document or worker that created it.

## Relevance to Vigil

- The `original-policy` / `originalPolicy` field in violation reports contains the full CSP header value, which uses the syntax and directives documented above.
- Every directive listed here can trigger violation reports that Vigil receives. The `effective-directive` / `effectiveDirective` field in reports will be one of these directive names.
- Understanding the fallback chain is important: a violation against `img-src` when only `default-src` is set will report `effective-directive: img-src` even though the policy only contains `default-src`.
- Source values help interpret why a violation was triggered: a `blocked-uri` that doesn't match any source expression in the effective directive causes the report.
- `'report-sample'` must be present in the directive for the `sample` / `script-sample` field to be populated in violation reports. Without it, Vigil will receive empty or absent sample data.
- The `report-to` directive (with `Reporting-Endpoints` header) sends reports in the modern Reporting API format, while `report-uri` sends reports in the legacy CSP format. Vigil supports both formats.
- Multiple policies mean a single page load can generate violation reports from different policies, each with a different `original-policy` value.

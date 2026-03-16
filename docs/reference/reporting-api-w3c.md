# Reporting API — W3C Working Draft

> Source: https://www.w3.org/TR/reporting-1/
> Spec Status: Working Draft (11 June 2025)
> Editors: Douglas Creager (GitHub), Ian Clelland (Google Inc.), Mike West (Google Inc.)
> Last Fetched: 2026-03-16

## Overview

The Reporting API defines a generic framework for browsers to deliver reports (such as CSP violations, deprecations, and interventions) to server-side collector endpoints. It establishes infrastructure for configuring named reporting endpoints via HTTP headers and provides a consistent delivery mechanism over HTTP POST.

## Report JSON Structure

Reports are delivered as a JSON array. Each element in the array is a report object with these fields:

| Field        | Type   | Description                                                                                     |
| ------------ | ------ | ----------------------------------------------------------------------------------------------- |
| `age`        | Number | Milliseconds between the report's timestamp and the time of delivery.                           |
| `type`       | String | The report type (e.g., `"csp-violation"`, `"deprecation"`).                                     |
| `url`        | String | The URL of the document or worker that generated the report. Username, password, and fragment are stripped for privacy. |
| `user_agent` | String | The value of the User-Agent header of the request from which the report was generated.           |
| `body`       | Object | Report-type-specific payload. For CSP violations, this is the `CSPViolationReportBody`.         |

The `url` field is privacy-sanitized: the spec mandates stripping username, password, and fragment components from the serialized URL to prevent capability URL leakage (though path-based sensitive information remains at risk).

The `age` field uses relative milliseconds rather than absolute timestamps to mitigate clock skew between browser and server.

### Example Report Payload

From the spec's Sample Reports section (three batched reports in a single POST):

```json
[
  {
    "type": "security-violation",
    "age": 10,
    "url": "https://example.com/vulnerable-page/",
    "user_agent": "Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0",
    "body": {
      "blocked": "https://evil.com/evil.js",
      "policy": "bad-behavior 'none'",
      "status": 200,
      "referrer": "https://evil.com/"
    }
  },
  {
    "type": "security-violation",
    "age": 20,
    "url": "https://example.com/vulnerable-page/",
    "user_agent": "Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0",
    "body": {
      "blocked": "https://evil.com/evil.css",
      "policy": "bad-behavior 'none'",
      "status": 200,
      "referrer": "https://evil.com/"
    }
  },
  {
    "type": "security-violation",
    "age": 30,
    "url": "https://example.com/vulnerable-page/",
    "user_agent": "Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0",
    "body": {
      "blocked": "https://evil.com/evil.html",
      "policy": "bad-behavior 'none'",
      "status": 200,
      "referrer": "https://evil.com/"
    }
  }
]
```

## Reporting-Endpoints Header

The `Reporting-Endpoints` HTTP response header is a Dictionary Structured Field. Each entry maps a name to an endpoint URL.

**Syntax:**

```http
Reporting-Endpoints: endpoint-1="https://example.com/reports"
```

**Multiple endpoints:**

```http
Reporting-Endpoints: csp-endpoint="https://example.com/csp-reports",
                     hpkp-endpoint="https://example.com/hpkp-reports"
```

**Integration with CSP:**

```http
Content-Security-Policy: ...; report-to csp-endpoint
Reporting-Endpoints: csp-endpoint="https://example.com/csp-reports"
```

**Rules:**

- Each entry value MUST be a string interpreted as a URI-reference.
- If the value is not a valid URI-reference, that endpoint member MUST be ignored.
- The URL MUST be potentially trustworthy (HTTPS or localhost).
- No parameters are defined for endpoints; any specified parameters are silently ignored.

## Delivery Algorithm

### Request Properties

Reports are sent using an HTTP request with these properties (from the spec's delivery algorithm in section 3.5.2):

| Property              | Value                       |
| --------------------- | --------------------------- |
| Method                | `POST`                      |
| Content-Type          | `application/reports+json`  |
| Request mode          | `cors`                      |
| Credentials mode      | `same-origin`               |
| Client                | `null`                      |
| Window                | `no-window`                 |
| Service-workers mode  | `none`                      |
| Initiator             | empty string                |
| Destination           | `report`                    |
| Unsafe-request flag   | set                         |

The request body is the JSON-serialized array of report objects.

### Batching Behavior

- The user agent periodically grabs the list of queued reports and delivers them to associated endpoints.
- The spec does **not** define a specific schedule — this is left to user agent discretion.
- Reports are grouped by endpoint first, then organized by origin within each endpoint.
- Delivery happens asynchronously per origin.
- A user agent SHOULD deliver reports as soon as possible after queuing, since report data is most useful shortly after generation.
- User agents MAY deliver only a subset of collected reports or endpoints (e.g., to conserve bandwidth).
- Multiple reports can be batched into a single POST request (as shown in the example above).

### Retry Behavior

The spec tracks a `failures` counter per endpoint. On delivery failure, the counter increments. The spec contains an open issue noting: "We don't specify any retry mechanism here for failed reports. We may want to add one here, or provide some indication that the delivery failed." In practice, browsers may retry, but the spec does not mandate it.

### Garbage Collection

User agents SHOULD periodically discard:

- **Stale endpoints**: endpoints whose `failures` count exceeds a user-agent-defined threshold (~5 is suggested as reasonable).
- **Old reports**: reports that have not been delivered within an arbitrary period (~2 days is suggested).

## CORS Requirements

Reports are sent with `mode: cors`. This means:

- **Cross-origin collectors** receive CORS-mode requests. The collector MUST handle CORS appropriately (respond to preflight `OPTIONS` requests with proper `Access-Control-Allow-Origin` and `Access-Control-Allow-Headers` headers, at minimum allowing `Content-Type`).
- **Same-origin collectors** receive credentials (`credentials: same-origin`), allowing session-aware processing.
- **Cross-origin collectors** do NOT receive credentials. The spec explicitly states this prevents leaking information to third-party endpoints that they could not obtain otherwise.

## HTTP Response Handling

The collector's HTTP response status code determines the outcome:

| Status Code | Result              | Effect                                                             |
| ----------- | ------------------- | ------------------------------------------------------------------ |
| 2xx         | **Success**         | Reports are accepted and removed from the browser's queue.         |
| 410 Gone    | **Remove Endpoint** | The endpoint is removed entirely; the browser stops sending to it. |
| Any other   | **Failure**         | The endpoint's `failures` counter increments. Reports remain queued for potential retry. |

**Collector guidance:**

- Return `200 OK` or `204 No Content` to acknowledge receipt.
- Return `410 Gone` only when the endpoint is intentionally decommissioned (this permanently stops reports from that browser).
- Avoid returning 4xx/5xx unless truly necessary, as it triggers the failure counter and may lead to the browser garbage-collecting the endpoint.

## Report Types

The Reporting API is type-agnostic; it provides the delivery framework. Report types are defined by other specifications:

| Type                 | Defined By                           | Notes                              |
| -------------------- | ------------------------------------ | ---------------------------------- |
| `csp-violation`      | Content Security Policy Level 3      | The primary type Vigil collects.   |
| `deprecation`        | Reporting API (built-in)             | Browser feature deprecation warnings. |
| `intervention`       | Reporting API (built-in)             | Browser interventions on page behavior. |
| `crash`              | Reporting API (built-in)             | Browser/tab crash reports.         |

Each type defines the schema of the `body` field. For CSP violations, the `body` contains fields like `documentURL`, `blockedURL`, `violatedDirective`, `effectiveDirective`, `originalPolicy`, `disposition`, `statusCode`, `referrer`, `sourceFile`, `lineNumber`, `columnNumber`, and `sample`.

## Relevance to Vigil

- Vigil's modern endpoint receives `application/reports+json` POST requests containing a JSON array of report objects.
- Reports arrive batched — multiple reports can be delivered in a single POST. Vigil must iterate the array and store each report individually.
- Vigil must return 2xx (ideally `204 No Content`) to acknowledge receipt. Returning `410 Gone` would permanently stop reports from that browser.
- CORS handling is needed if Vigil runs on a different origin than the reporting sites. This means responding to `OPTIONS` preflight requests with appropriate `Access-Control-Allow-Origin`, `Access-Control-Allow-Methods: POST`, and `Access-Control-Allow-Headers: Content-Type` headers.
- The `body` field contains the type-specific report data. For CSP reports (`type: "csp-violation"`), this is the `CSPViolationReportBody`.
- The `age` field indicates milliseconds since the violation occurred, which is useful for timing analysis and detecting delayed report delivery. Since it is relative (not an absolute timestamp), it avoids clock skew issues between browser and server.
- The `url` field has username, password, and fragment stripped for privacy — Vigil should not expect to receive these components.
- Browsers may garbage-collect endpoints after ~5 consecutive failures and discard undelivered reports after ~2 days. Vigil's availability directly impacts whether reports are received.
- Credentials are only sent for same-origin requests. If Vigil is cross-origin, it will not receive cookies or other credentials.

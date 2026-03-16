# CSP Standards Reference Documents

Authoritative reference documents for CSP violation report formats, delivery semantics, and field definitions. These are curated extractions from W3C specs and MDN documentation, focused on what's relevant to Vigil as a CSP report collector.

## W3C Specifications

| Document | Source | Description |
|----------|--------|-------------|
| [csp-level3-w3c.md](csp-level3-w3c.md) | [W3C CSP Level 3](https://www.w3.org/TR/CSP3/) | Current working draft. Defines CSPViolationReportBody, report generation algorithm, strip-url algorithm, blocked-uri special values, report-uri vs report-to interaction. |
| [csp-level2-w3c.md](csp-level2-w3c.md) | [W3C CSP Level 2](https://www.w3.org/TR/CSP2/) | Recommendation (2016). Defines legacy `csp-report` JSON wrapper format, all kebab-case field names, `application/csp-report` Content-Type, cross-origin URI truncation. |
| [reporting-api-w3c.md](reporting-api-w3c.md) | [W3C Reporting API](https://www.w3.org/TR/reporting-1/) | Defines the generic report delivery mechanism: JSON array format, `application/reports+json` Content-Type, batching, CORS, HTTP response handling (2xx/410), Reporting-Endpoints header. |

## MDN References

| Document | Source | Description |
|----------|--------|-------------|
| [mdn-csp-violation-report-body.md](mdn-csp-violation-report-body.md) | [MDN CSPViolationReportBody](https://developer.mozilla.org/en-US/docs/Web/API/CSPViolationReportBody) | Complete property list for modern report body (camelCase fields), types, JSON serialization, field name mapping between modern and legacy formats. |
| [mdn-csp-report-uri.md](mdn-csp-report-uri.md) | [MDN report-uri](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/report-uri) | Legacy `report-uri` directive. Full `csp-report` JSON format, all kebab-case fields, `application/csp-report` Content-Type, blocked-uri truncation rules, security considerations. |
| [mdn-csp-report-to.md](mdn-csp-report-to.md) | [MDN report-to](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/report-to) | Modern `report-to` directive. Interaction with Reporting-Endpoints header, report format, migration from report-uri, browser compatibility. |
| [mdn-csp-guide.md](mdn-csp-guide.md) | [MDN CSP Guide](https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CSP) | CSP overview. All directives by category, source values, Report-Only mode, meta tag limitations, violation report lifecycle, blocked-uri edge cases. |
| [mdn-reporting-endpoints.md](mdn-reporting-endpoints.md) | [MDN Reporting-Endpoints](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Reporting-Endpoints) | Reporting-Endpoints HTTP header. Syntax, endpoint naming, HTTPS requirement, replaces deprecated Report-To header. |
| [mdn-csp-header.md](mdn-csp-header.md) | [MDN Content-Security-Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy) | Complete directive list with descriptions, source list values, special keywords. |
| [mdn-csp-report-only.md](mdn-csp-report-only.md) | [MDN CSP-Report-Only](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy-Report-Only) | Report-Only mode. Differences from enforcing CSP, disposition field, usage with report-to/report-uri. |

## Field Name Cross-Reference

Vigil handles both legacy and modern CSP report formats. The field names differ:

| Legacy (csp-report) | Modern (CSPViolationReportBody) | Vigil gjson Path (Legacy) | Vigil gjson Path (Modern) |
|---------------------|--------------------------------|--------------------------|--------------------------|
| `document-uri` | `documentURL` | `csp-report.document-uri` | `body.documentURL` |
| `blocked-uri` | `blockedURL` | `csp-report.blocked-uri` | `body.blockedURL` |
| `effective-directive` | `effectiveDirective` | `csp-report.effective-directive` | `body.effectiveDirective` |
| `violated-directive` | _(removed in modern)_ | `csp-report.violated-directive` | — |
| `original-policy` | `originalPolicy` | `csp-report.original-policy` | `body.originalPolicy` |
| `referrer` | `referrer` | `csp-report.referrer` | `body.referrer` |
| `status-code` | `statusCode` | `csp-report.status-code` | `body.statusCode` |
| `source-file` | `sourceFile` | `csp-report.source-file` | `body.sourceFile` |
| `line-number` | `lineNumber` | `csp-report.line-number` | `body.lineNumber` |
| `column-number` | `columnNumber` | `csp-report.column-number` | `body.columnNumber` |
| `script-sample` | `sample` | `csp-report.script-sample` | `body.sample` |
| `disposition` | `disposition` | `csp-report.disposition` | `body.disposition` |
| — | — | — | `user_agent` |

## Content-Type Cross-Reference

| Format | Content-Type | Spec |
|--------|-------------|------|
| Legacy (report-uri) | `application/csp-report` | CSP Level 2 |
| Modern (report-to) | `application/reports+json` | Reporting API |

package sample

import (
	"math/rand/v2"
)

var directives = []string{
	"script-src-elem", "style-src-elem", "img-src",
	"font-src", "connect-src", "default-src",
	"frame-src", "media-src", "object-src",
}

var dispositions = []string{"enforce", "report"}

var blockedURIs = []string{
	"https://cdn.example.com/script.js",
	"https://tracker.example.net/pixel.gif",
	"https://ads.example.org/banner.js",
	"inline", "data:", "blob:",
}

var documentURIs = []string{
	"https://example.com/",
	"https://example.com/login",
	"https://example.com/dashboard",
	"https://example.com/settings",
	"https://example.com/checkout",
}

var userAgents = []string{
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:128.0) Gecko/20100101 Firefox/128.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36 Edg/127.0.0.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15",
}

var statusCodes = []int{200, 0}

func pick[T any](items []T) T {
	return items[rand.IntN(len(items))]
}

// GenerateLegacy returns a CSP report in the legacy csp-report format.
func GenerateLegacy(random bool) map[string]any {
	if !random {
		return map[string]any{
			"csp-report": map[string]any{
				"blocked-uri":         "https://cdn.example.com/script.js",
				"disposition":         "enforce",
				"document-uri":        "https://example.com/page",
				"effective-directive": "script-src-elem",
				"original-policy":     "default-src 'self'; script-src 'self'; report-uri /report",
				"referrer":            "https://example.com/",
				"status-code":         200,
				"violated-directive":  "script-src-elem",
				"source-file":         "https://cdn.example.com/script.js",
				"line-number":         42,
				"column-number":       15,
			},
		}
	}

	directive := pick(directives)
	return map[string]any{
		"csp-report": map[string]any{
			"blocked-uri":         pick(blockedURIs),
			"disposition":         pick(dispositions),
			"document-uri":        pick(documentURIs),
			"effective-directive": directive,
			"original-policy":     "default-src 'self'; " + directive + " 'self'; report-uri /report",
			"referrer":            pick(documentURIs),
			"status-code":         pick(statusCodes),
			"violated-directive":  directive,
			"source-file":         pick(blockedURIs),
			"line-number":         rand.IntN(500) + 1,
			"column-number":       rand.IntN(100) + 1,
		},
	}
}

// GenerateModern returns a CSP report in the modern Reporting API format.
func GenerateModern(random bool) map[string]any {
	if !random {
		return map[string]any{
			"age": 500,
			"body": map[string]any{
				"blockedURL":         "https://cdn.example.com/script.js",
				"columnNumber":       15,
				"disposition":        "enforce",
				"documentURL":        "https://example.com/page",
				"effectiveDirective": "script-src-elem",
				"lineNumber":         42,
				"originalPolicy":     "default-src 'self'; script-src 'self'; report-to csp-endpoint",
				"referrer":           "https://example.com/",
				"sample":             "console.log('test')",
				"sourceFile":         "https://cdn.example.com/script.js",
				"statusCode":         200,
			},
			"type":       "csp-violation",
			"url":        "https://example.com/page",
			"user_agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36",
		}
	}

	directive := pick(directives)
	docURL := pick(documentURIs)
	return map[string]any{
		"age": rand.IntN(86401),
		"body": map[string]any{
			"blockedURL":         pick(blockedURIs),
			"columnNumber":       rand.IntN(100) + 1,
			"disposition":        pick(dispositions),
			"documentURL":        docURL,
			"effectiveDirective": directive,
			"lineNumber":         rand.IntN(500) + 1,
			"originalPolicy":     "default-src 'self'; " + directive + " 'self'; report-to csp-endpoint",
			"referrer":           pick(documentURIs),
			"sample":             "console.log('test')",
			"sourceFile":         pick(blockedURIs),
			"statusCode":         pick(statusCodes),
		},
		"type":       "csp-violation",
		"url":        docURL,
		"user_agent": pick(userAgents),
	}
}

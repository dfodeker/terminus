// internal/http/middleware/versioning.go
package middleware

import (
	"context"
	"net/http"
	"regexp"
)

var versionRegex = regexp.MustCompile(`^/api/(\d{4}-\d{2})`)

// Supported API versions
var SupportedVersions = map[string]bool{
	"2024-01": true,
	"2024-07": true,
	"2025-01": true, // Current
}

var LatestVersion = "2025-01"

const APIVersionKey contextKey = "api_version"

func APIVersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract version from path
		matches := versionRegex.FindStringSubmatch(r.URL.Path)

		var version string
		if len(matches) >= 2 {
			version = matches[1]
		} else {
			// Check header fallback
			version = r.Header.Get("X-API-Version")
			if version == "" {
				version = LatestVersion
			}
		}

		if !SupportedVersions[version] {
			http.Error(w, "unsupported API version", http.StatusBadRequest)
			return
		}

		// Add version to response headers
		w.Header().Set("X-API-Version", version)

		// Add deprecation warning for old versions
		if version != LatestVersion {
			w.Header().Set("X-API-Deprecation-Warning",
				"This API version is deprecated. Please upgrade to "+LatestVersion)
		}

		ctx := context.WithValue(r.Context(), APIVersionKey, version)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// limit the amount of request data the web server will process
package web

import (
	"net/http"
)

// BodyLimitSize is the largest amount of bytes the body of the request
// is permitted to accept
var BodyLimitSize int64 = 1 << 17 // ~125k

// bodyLimitMiddleware limits request bodies
func bodyLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, BodyLimitSize) // ~125k
		next.ServeHTTP(w, r)
	})
}

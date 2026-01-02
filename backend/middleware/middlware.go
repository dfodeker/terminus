package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

type ctxKey int

const requestIDKey ctxKey = iota

const HeaderRequestID = "X-Request-Id"

func GetRequestID(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}

//request id endures everyrequest has a request id
//if the client sends x request id , we accept it
//other wise generate a new one

func newReqID() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		//supposed to be pretty rare, fallback to something else
		//can swap to be time and a counter later
		return "reqid-rand-failed"

	}
	return hex.EncodeToString(b[:])
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := sanitizeReqID(r.Header.Get(HeaderRequestID))
		if rid == "" {
			rid = newReqID()
		}

		// Add to context and echo in response.
		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		w.Header().Set(HeaderRequestID, rid)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sanitizeReqID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) > 128 {
		s = s[:128]
	}
	s = strings.Map(func(r rune) rune {
		if r <= 31 || r == 127 {
			return -1
		}
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, s)
	return s
}

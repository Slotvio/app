package middleware

import (
	"net/http"

	requestid "github.com/aleksandrboro/slotvio/internal/request_id"
	"github.com/google/uuid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.NewV7()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal error"))
			return
		}

		requestID := id.String()
		w.Header().Set("X-Request-ID", requestID)

		ctx := requestid.WithRequestID(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

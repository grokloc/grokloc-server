package server

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// RequestLogger is a middleware that adds logging to each request
func (srv *Instance) RequestLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer zap.L().Sync() // nolint
		sugar := zap.L().Sugar()
		sugar.Infow("request",
			"reqid", middleware.GetReqID(ctx),
			"method", r.Method,
			"path", r.URL,
			"remote", r.RemoteAddr,
			"headers", r.Header,
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

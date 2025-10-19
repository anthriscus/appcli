package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/anthriscus/appcli/appcontext"
)

func addMiddleware(mux *http.ServeMux) http.HandlerFunc {
	//muxChain := tracerMiddleware(contentTypeMiddleware(mux))
	// potential for more middleware wrappers here
	// but we apply contentType only to REST calls so don't add it to all pathways in here
	muxChain := tracerMiddleware(mux)
	return muxChain
}

func tracerMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("ROUTE path: %s\n", r.URL.Path)
		id := appcontext.GenerateId()
		ctx := context.WithValue(r.Context(), appcontext.TraceIdKey, id)
		logger.Log.InfoContext(ctx, "tracer", "route", r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// only for api calls and not general file content
func contentTypeMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log.InfoContext(r.Context(), "contentType set", "route", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

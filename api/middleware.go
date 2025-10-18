package api

import (
	"fmt"
	"net/http"
)

// func addMiddleware(mux *http.ServeMux, l logging.AppLogger) http.HandlerFunc {
func addMiddleware(mux *http.ServeMux) http.HandlerFunc {
	//muxChain := tracerMiddleware(contentTypeMiddleware(mux))
	muxChain := tracerMiddleware(mux)
	return muxChain
}

func tracerMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("ROUTE path: %s\n", r.URL.Path)
		serverLogger.Log.Info("tracer", "route", r.URL.Path)
		next.ServeHTTP(w, r)
		// TODO get the ctx and add some context tracing info in here
		// and log it to the global logger
		// next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// only for api calls and not general file content
func contentTypeMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverLogger.Log.Info("contentType", "route", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

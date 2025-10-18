package api

import (
	"encoding/json"
	"net/http"
)

const (
	StoreDescription string = "to-do app which allows users to create a list of to-do tasks."
)

type AboutInfo struct {
	Description string
}

// help about
func AboutJson(w http.ResponseWriter, r *http.Request) {
	s := aboutJson()
	if ok := json.NewEncoder(w).Encode(&s); ok != nil {
		// cannot write htp error after write, so log
	}
}
func aboutJson() AboutInfo {
	return AboutInfo{Description: StoreDescription}
}
func About(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "./api/template/index.html")
}

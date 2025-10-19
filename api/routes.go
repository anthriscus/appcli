package api

import (
	"net/http"
	"os"
)

// type apiHandler func(w http.ResponseWriter, r *http.Request)

type route struct {
	method  string // separated from route for clarity
	route   string
	handler http.HandlerFunc //apiHandler
	isweb   bool             // flag type of endpoint to distinguish between api/webpage content
}

var Routes = []route{}

func addRoutes(mux *http.ServeMux) {
	Routes = []route{
		{method: "DELETE", route: "/delete/{taskId}", handler: Delete},
		{method: "GET", route: "/aboutapi", handler: AboutJson},
		{method: "GET", route: "/about", handler: About, isweb: true},
		{method: "GET", route: "/get/{taskId}", handler: GetByIndex},
		{method: "GET", route: "/get", handler: GetList},
		{method: "POST", route: "/create", handler: Create},
		{method: "PUT", route: "/update", handler: UpdateTask},
	}

	if pth, ok := os.Getwd(); ok == nil {
		fs := http.FileServer(http.Dir(pth + "\\files"))
		mux.Handle("GET"+" "+"/", fs)
	}
	// the api routes have a json media type header
	for _, r := range Routes {
		mux.HandleFunc(r.method+" "+r.route, contentTypeMiddleware(r.handler))
	}
}

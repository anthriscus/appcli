package api

import (
	"encoding/json" // only temp in this package for our mock data which later will be removed and will become a byte stream
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/anthriscus/appcli/logging"
	"github.com/anthriscus/appcli/store"
)

const (
	applicationHttpPort int = 8080 // maybe get from env (which case would be a string...)
)

type apiError struct {
	Error string
}

var (
	serverLogger logging.AppLogger
)

func Run(l logging.AppLogger) {
	fmt.Println("hello from api module")
	serverLogger = l
	l.Log.Info("hello from api module")

	// id := store.GenerateId()
	// ctx := context.WithValue(context.Background(), appcontext.TraceIdKey, id)

	mux := http.NewServeMux()
	addRoutes(mux)
	listenOn := strconv.Itoa(applicationHttpPort)
	slog.Info("Starting server on :" + listenOn)
	endPoint := ":" + listenOn

	muxChain := addMiddleware(mux)
	if err := http.ListenAndServe(endPoint, muxChain); err != nil {
		slog.Error("Http Server failed", "error", err)
	}
}

func Create(w http.ResponseWriter, r *http.Request) {
	var item store.TodoListItem
	if ok := json.NewDecoder(r.Body).Decode(&item); ok != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	} else {
		if newItem, ok := store.Create(item); ok != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonError(ok))
		} else {
			w.WriteHeader((http.StatusOK))
			if ok := json.NewEncoder(w).Encode(&newItem); ok != nil {
				// http.Error(w, fmt.Sprint(ok), http.StatusBadRequest)
			}
		}
	}
}

func GetByIndex(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("taskId")
	if id, ok := strconv.Atoi(taskId); ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		if item, ok := store.GetByIndex(id); ok != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			if ok := json.NewEncoder(w).Encode(&item); ok != nil {
				// cannot write htp error after write, so log
				// http.Error(w, "failed to fetch task", http.StatusBadRequest)
				return
			}
		}
	}
}

func GetList(w http.ResponseWriter, r *http.Request) {
	if items, ok := store.GetList(); ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		if ok := json.NewEncoder(w).Encode(&items); ok != nil {
			// cannot write htp error after write
			// http.Error(w, "failed to fetch task", http.StatusBadRequest)
			return
		}
	}
}

func UpdateByIndex(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("taskId")
	if id, ok := strconv.Atoi(taskId); ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		var item store.TodoListItem
		if ok := json.NewDecoder(r.Body).Decode(&item); ok != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		} else {
			if newItem, ok := store.UpdateByIndex(id, item); ok != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(jsonError(ok))
			} else {
				if ok := json.NewEncoder(w).Encode(&newItem); ok != nil {
					// http.Error(w, "failed to update task", http.StatusBadRequest)
				}
			}
		}
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("taskId")
	if id, ok := strconv.Atoi(taskId); ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		if ok := store.Delete(id); ok != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonError(ok))
		} else {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
}

func jsonError(err error) []byte {
	e := apiError{Error: err.Error()}
	return []byte(fmt.Sprintf("%+v", e))
}

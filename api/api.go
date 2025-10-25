package api

import (
	"context"
	"encoding/json" // only temp in this package for our mock data which later will be removed and will become a byte stream
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/anthriscus/appcli/appcontext"
	"github.com/anthriscus/appcli/logging"
	"github.com/anthriscus/appcli/store"
)

const (
	applicationHttpPort int = 8080 // maybe get from env (which case would be a string...)
)

type apiError struct {
	Error string
}

func Run() {
	id := appcontext.GenerateId()
	ctx := context.WithValue(context.Background(), appcontext.TraceIdKey, id)

	mux := http.NewServeMux()
	addRoutes(mux)
	listenOn := strconv.Itoa(applicationHttpPort)
	endPoint := ":" + listenOn
	muxChain := addMiddleware(mux)

	logging.Log().InfoContext(ctx, "Starting server", "listeningOn", listenOn)
	fmt.Printf("Starting server listening on:%s\n ", listenOn)

	// await on a shutdown
	shutdownChan := make(chan os.Signal, 1)
	// define what we will wait for
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    endPoint,
		Handler: muxChain,
	}

	// start server on a routine so we can catch the shutdownChan cancel below.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logging.Log().ErrorContext(ctx, "Listening ended", "error", err)
		}
	}()

	// spin up the actor
	go func() {
		Actor()
	}()

	// block until the signal
	fmt.Println("waiting for your signal")
	sig := <-shutdownChan
	handleShutdown(ctx, srv, sig)
}

func handleShutdown(ctx context.Context, srv *http.Server, sig os.Signal) {
	fmt.Printf("Got cancel signal %+v\n", sig)
	logging.Log().InfoContext(ctx, "Shutdown requested", "signal", sig)

	// shutdown the server
	if ok := srv.Shutdown(ctx); ok != nil {
		logging.Log().ErrorContext(ctx, "Shutdown server failed", "error", ok)
	} else {
		logging.Log().InfoContext(ctx, "Server shutdown")
	}

	// commit the data
	fmt.Println("Commiting data in shutdown")
	logging.Log().InfoContext(ctx, "Commiting data in shutdown")
	if ok := store.Commit(ctx); ok != nil {
		logging.Log().ErrorContext(ctx, "Data commit failed", "error", ok)
	} else {
		logging.Log().InfoContext(ctx, "Committed")
	}
	fmt.Println("Goodbye")
	logging.Log().InfoContext(ctx, "Goodbye")
}

func Create(w http.ResponseWriter, r *http.Request) {
	var item store.TodoListItem
	if ok := json.NewDecoder(r.Body).Decode(&item); ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(jsonError(fmt.Errorf("invalid json")))
		return
	} else {
		actorHandler(apiCreate(StoreRequest{writer: w, request: r, todoListItem: item}))
		result := <-ResponseChan
		if result.err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jsonError(result.err))
			// if newItem, ok := store.Create(r.Context(), item); ok != nil {
			// w.WriteHeader(http.StatusBadRequest)
			// json.NewEncoder(w).Encode(jsonError(ok))
		} else {
			w.WriteHeader((http.StatusCreated))
			// if ok := json.NewEncoder(w).Encode(&newItem); ok != nil {
			if ok := json.NewEncoder(w).Encode(&result.todoListItem); ok != nil {
				logging.Log().ErrorContext(r.Context(), "Create", "error", ok)
			}
		}
	}
}

// shows use of index in REST path
func GetByIndex(w http.ResponseWriter, r *http.Request) {
	//
	taskId := r.PathValue("taskId")
	if id, ok := strconv.Atoi(taskId); ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(jsonError(ok))
		return
	} else {
		findData := store.TodoListItem{Line: id}
		actorHandler(apiGetListByIndex(StoreRequest{writer: w, request: r, todoListItem: findData}))
		result := <-ResponseChan
		// if item, ok := store.GetByIndex(id); ok != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	json.NewEncoder(w).Encode(jsonError(ok))
		// 	return
		// } else {
		// 	if ok := json.NewEncoder(w).Encode(&item); ok != nil {
		// 		logging.Log().ErrorContext(r.Context(), "GetByIndex", "error", ok)
		// 		return
		// 	}
		// }
		//
		if result.err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jsonError(ok))
			return
		} else {
			if ok := json.NewEncoder(w).Encode(&result.todoListItem); ok != nil {
				logging.Log().ErrorContext(r.Context(), "GetByIndex", "error", ok)
				return
			}
		}
		//
	}
}

func GetList(w http.ResponseWriter, r *http.Request) {
	actorHandler(apiGetList())
	result := <-ResponseChan
	// if items, ok := store.GetList(); ok != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	json.NewEncoder(w).Encode(jsonError(ok))
	// 	return
	// } else {
	// 	if ok := json.NewEncoder(w).Encode(&items); ok != nil {
	// 		logging.Log().ErrorContext(r.Context(), "GetList", "error", ok)
	// 		return
	// 	}
	// }
	if result.err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(jsonError(result.err))
		return
	} else {
		if ok := json.NewEncoder(w).Encode(&result.todoListItems); ok != nil {
			logging.Log().ErrorContext(r.Context(), "GetList", "error", result.err)
			return
		}
	}
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	var item store.TodoListItem
	if ok := json.NewDecoder(r.Body).Decode(&item); ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(jsonError(fmt.Errorf("invalid json")))
		return
	} else {
		// if newItem, ok := store.Update(r.Context(), item); ok != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	json.NewEncoder(w).Encode(jsonError(ok))
		// } else {
		// 	if ok := json.NewEncoder(w).Encode(&newItem); ok != nil {
		// 		logging.Log().ErrorContext(r.Context(), "UpdateTask", "error", ok)
		// 	}
		// }
		actorHandler(apiUpdate(StoreRequest{writer: w, request: r, todoListItem: item}))
		result := <-ResponseChan
		if result.err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jsonError(ok))
		} else {
			if ok := json.NewEncoder(w).Encode(&result.todoListItem); ok != nil {
				logging.Log().ErrorContext(r.Context(), "UpdateTask", "error", ok)
			}
		}
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("taskId")
	if id, ok := strconv.Atoi(taskId); ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(jsonError(fmt.Errorf("bad taskid")))
		return
	} else {
		// if ok := store.Delete(r.Context(), id); ok != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	json.NewEncoder(w).Encode(jsonError(ok))
		// } else {
		// 	w.WriteHeader(http.StatusNoContent)
		// 	return
		// }
		deleteData := store.TodoListItem{Line: id}
		actorHandler(apiDelete(StoreRequest{writer: w, request: r, todoListItem: deleteData}))
		result := <-ResponseChan
		if result.err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jsonError(ok))
		} else {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
}

func GetActiveList(w http.ResponseWriter, r *http.Request) {
	if items, ok := store.GetList(); ok != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		if t, ok := template.ParseFiles("./api/template/activetodolist.html"); ok != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			// compile the template and serve to the response writer
			if ok := t.Execute(w, items); ok != nil {
				logging.Log().ErrorContext(r.Context(), "template parse failed")
			}
		}
	}
}

// helper format for json in response results
func jsonError(err error) apiError {
	return apiError{Error: err.Error()}
}

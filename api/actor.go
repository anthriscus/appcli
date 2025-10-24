package api

import (
	"fmt"
	"net/http"

	"github.com/anthriscus/appcli/store"
)

type StoreRequest struct {
	writer       http.ResponseWriter // needed ? may be expand the go func to write to it/ finish the request
	request      *http.Request
	todoListItem store.TodoListItem
}

type StoreResult struct {
	todoListItems store.TodoListItems
	todoListItem  store.TodoListItem
	err           error
}

// runs something and gets a result
type actorRunner func() StoreResult

// receives a runner and channel to write the result back to
type RequestChannel struct {
	runner          actorRunner
	responseChannel *chan StoreResult // where we want the results to be piped back to
}

var (
	RequestsPipeline = make(chan RequestChannel) // pipe data/func to store actions
	ResponsePipeline = make(chan StoreResult)    // pipe data from store actions
)

func actorHandler(handler actorRunner) {
	var request RequestChannel
	request.runner = handler
	request.responseChannel = &ResponsePipeline
	RequestsPipeline <- request
	fmt.Println("Actor pushed results to request pipeline channel")
}

func Actor() {
	fmt.Printf("Actor started with channel size of %d\n", cap(RequestsPipeline))
	for req := range RequestsPipeline {
		result := req.runner()
		*req.responseChannel <- result
		fmt.Println("Actor pushed results to response pipeline channel")
	}
}

var apiCreate = func(storeRequest StoreRequest) actorRunner {
	return func() StoreResult {
		newItem, ok := store.Create(storeRequest.request.Context(), storeRequest.todoListItem)
		return StoreResult{
			todoListItem:  newItem,
			err:           ok,
			todoListItems: store.TodoListItems{},
		}
	}
}

var apiUpdate = func(storeRequest StoreRequest) actorRunner {
	return func() StoreResult {
		newItem, ok := store.Update(storeRequest.request.Context(), storeRequest.todoListItem)
		return StoreResult{
			todoListItem: newItem,
			err:          ok,
		}
	}
}

var apiDelete = func(storeRequest StoreRequest) actorRunner {
	return func() StoreResult {
		ok := store.Delete(storeRequest.request.Context(), storeRequest.todoListItem.Line)
		return StoreResult{
			todoListItem: store.TodoListItem{},
			err:          ok,
		}
	}
}

var apiGetList = func() actorRunner {
	return func() StoreResult {
		items, ok := store.GetList()
		return StoreResult{
			todoListItem:  store.TodoListItem{},
			todoListItems: items,
			err:           ok,
		}
	}
}

var apiGetListByIndex = func(storeRequest StoreRequest) actorRunner {
	return func() StoreResult {
		item, ok := store.GetByIndex(storeRequest.todoListItem.Line)
		return StoreResult{
			todoListItem: item,
			err:          ok,
		}
	}
}

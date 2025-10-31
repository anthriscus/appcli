package api

import (
	"context"
	"fmt"

	"github.com/anthriscus/appcli/store"
)

type StoreRequest struct {
	ctx          context.Context
	todoListItem store.TodoListItem
}

type StoreResult struct {
	todoListItems store.TodoListItems
	todoListItem  store.TodoListItem
	err           error
}

// runs something and gets a result
type actorCommand func() StoreResult

// receives a command and channel to write the result back to
type RequestChannel struct {
	command         actorCommand
	responseChannel *chan StoreResult // where we want the results to be piped back to
}

var (
	RequestsChan = make(chan RequestChannel) // channel for commands that call store actions
	ResponseChan = make(chan StoreResult)    // channel for returning result from store actions
)

func actorHandler(handler actorCommand, respChan chan StoreResult) {
	var request RequestChannel
	request.command = handler
	request.responseChannel = &respChan
	RequestsChan <- request
	fmt.Println("Actor pushed results to request channel")
}

func Actor() {
	fmt.Printf("Actor started with channel size of %d\n", cap(RequestsChan))
	for req := range RequestsChan {
		go func() {
			result := req.command()
			// return the result of the command in the returning channel
			*req.responseChannel <- result
			fmt.Println("Actor pushed results to response channel")
		}()
	}
}

var apiCreate = func(storeRequest StoreRequest) actorCommand {
	return func() StoreResult {
		newItem, ok := store.Create(storeRequest.ctx, storeRequest.todoListItem)
		return StoreResult{
			todoListItem:  newItem,
			err:           ok,
			todoListItems: store.TodoListItems{},
		}
	}
}

var apiUpdate = func(storeRequest StoreRequest) actorCommand {
	return func() StoreResult {
		newItem, ok := store.Update(storeRequest.ctx, storeRequest.todoListItem)
		return StoreResult{
			todoListItem: newItem,
			err:          ok,
		}
	}
}

var apiDelete = func(storeRequest StoreRequest) actorCommand {
	return func() StoreResult {
		ok := store.Delete(storeRequest.ctx, storeRequest.todoListItem.Line)
		return StoreResult{
			todoListItem: store.TodoListItem{},
			err:          ok,
		}
	}
}

var apiGetList = func() actorCommand {
	return func() StoreResult {
		items, ok := store.GetList()
		return StoreResult{
			todoListItem:  store.TodoListItem{},
			todoListItems: items,
			err:           ok,
		}
	}
}

var apiGetListByIndex = func(storeRequest StoreRequest) actorCommand {
	return func() StoreResult {
		item, ok := store.GetByIndex(storeRequest.todoListItem.Line)
		return StoreResult{
			todoListItem: item,
			err:          ok,
		}
	}
}

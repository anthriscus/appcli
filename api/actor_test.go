package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/anthriscus/appcli/api/internal"
	"github.com/anthriscus/appcli/logging"
	"github.com/anthriscus/appcli/store"
)

func TestMain(m *testing.M) {
	logging.Default()
	// load the sample todolists into memory
	dataFile, _ := filepath.Abs("../testdata/todolist.500.json")
	ctx := context.Background()
	store.OpenSession(ctx, dataFile)

	// startup the api actor to open the channels
	go func() {
		Actor()
	}()
	m.Run()
}

// here we mock the server call
// for testing the actor concurrency on the get from store
func TestGetList(t *testing.T) {

	var tests = struct {
		runRequests int
		want        int
	}{runRequests: 100, want: 10}

	t.Parallel()

	var wg sync.WaitGroup
	for requestStep := range tests.runRequests {
		wg.Add(1)
		go func(clientId int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/get", nil)
			w := httptest.NewRecorder()
			GetList(w, req)
			res := w.Result()
			fetchedItems := decodeJsonBodyItems(res)
			wantCountAtleast := tests.want
			if fetchedItems == nil {
				t.Errorf("ClientId %d empty list returned, wanted a list", clientId)
			} else if len(fetchedItems) <= wantCountAtleast {
				t.Errorf("ClientId %d list returned %d, wanted a list of at least %d", clientId, len(fetchedItems), wantCountAtleast)
			} else {
				t.Logf("ClientId %d list returned %d ", clientId, len(fetchedItems))
			}
		}(requestStep)
	}
	// wait for all client runs to end
	wg.Wait()
}

func TestAdd(t *testing.T) {
	apiCreate := "/create"
	numClients := 100
	results := make(chan string, numClients)
	t.Parallel()

	var wg sync.WaitGroup
	for i := range numClients {
		wg.Add(1)
		go func(clientId int) {
			defer wg.Done()

			candidates := internal.GenerateDummyTasks(1)
			candidate := candidates[0]
			// add info for id in description
			candidate.Description = fmt.Sprintf("client %d ,%x", clientId, candidate.Description)
			jsonData, _ := encodeJsonBodyItem(candidate)
			payload := bytes.NewBuffer(jsonData)
			req := httptest.NewRequest(http.MethodPost, apiCreate, payload)
			w := httptest.NewRecorder()
			t.Logf("running add task for Client:%d", clientId)
			Create(w, req)

			if w.Result().StatusCode != http.StatusCreated {
				t.Errorf("clientID %d, create failed, wanted: %d got:%d", clientId, http.StatusCreated, w.Result().StatusCode)
			} else if w.Body == nil {
				t.Error("create failed, bad body returned")
			}
			createdTodo := decodeJsonRecorderBodyItem(w.Body.Bytes())
			// assert fields here
			if createdTodo.Description != candidate.Description {
				t.Error("create failed, expected descriptions to match")
			}
			// push these to results chan
			results <- fmt.Sprintf("clientID %d, created new task item with line id:%d", clientId, createdTodo.Line)
		}(i)
	}
	// wait for all client runs to end
	wg.Wait()
	close(results)
	t.Logf("Adding tasks. Found %d items in Client results", len(results))
	for result := range results {
		t.Log(result)
	}
}

// update the same line numbers 1..100 with different runnning client ids
func TestUpdate(t *testing.T) {
	apiCreate := "/update"

	lineNumbers := 100
	numClients := 10
	results := make(chan string, numClients*lineNumbers)

	t.Parallel()

	var wg sync.WaitGroup
	// lineNumber is the line item we are updating
	for i := range numClients {
		wg.Add(1)
		go func(clientId int) {
			defer wg.Done()
			for j := range lineNumbers {
				lineNumber := j + 1
				candidates := internal.GenerateDummyTasks(1)
				candidate := candidates[0]
				// add info for id in description
				candidate.Description = fmt.Sprintf("line %d ,%x", lineNumber, candidate.Description)
				// the line that we are updating
				candidate.Line = lineNumber
				jsonData, _ := encodeJsonBodyItem(candidate)
				payload := bytes.NewBuffer(jsonData)
				req := httptest.NewRequest(http.MethodPut, apiCreate, payload)
				w := httptest.NewRecorder()
				t.Logf("running update task for Client:%d, line number %d", clientId, lineNumber)
				UpdateTask(w, req)

				if w.Result().StatusCode != http.StatusOK {
					t.Errorf("client id %d Line %d, update failed, wanted: %d got:%d", clientId, lineNumber, http.StatusOK, w.Result().StatusCode)
				} else if w.Body == nil {
					t.Error("update failed, bad body returned")
				}
				createdTodo := decodeJsonRecorderBodyItem(w.Body.Bytes())
				// assert fields here
				if createdTodo.Description != candidate.Description {
					t.Error("update failed, expected descriptions to match")
				}
				// push these to results chan
				results <- fmt.Sprintf("Client id %d Line %d, updated task item with line id:%d", clientId, lineNumber, createdTodo.Line)
			}
		}(i)
	}
	// wait for all client runs to end
	wg.Wait()
	close(results)
	t.Logf("Adding tasks. Found %d items in Client results", len(results))
	for result := range results {
		t.Log(result)
	}
}

func encodeJsonBodyItem(todoItem store.TodoListItem) ([]byte, error) {
	data, err := json.Marshal(todoItem)
	return data, err
}

func decodeJsonBodyItems(resp *http.Response) store.TodoListItems {
	var todoItems store.TodoListItems

	if bytes, ok := io.ReadAll(resp.Body); ok != nil {
		return store.TodoListItems{}
	} else {
		if ok := json.Unmarshal(bytes, &todoItems); ok != nil {
			return store.TodoListItems{}
		} else {
			return todoItems
		}
	}
}

func decodeJsonRecorderBodyItem(body []byte) store.TodoListItem {
	var todoItem store.TodoListItem
	if ok := json.Unmarshal(body, &todoItem); ok != nil {
		return store.TodoListItem{}
	} else {
		return todoItem
	}
}

// func decodeJsonBodyItem(resp *http.Response) store.TodoListItem {
// 	var todoItem store.TodoListItem

// 	if bytes, ok := io.ReadAll(resp.Body); ok != nil {
// 		return store.TodoListItem{}
// 	} else {
// 		if ok := json.Unmarshal(bytes, &todoItem); ok != nil {
// 			return store.TodoListItem{}
// 		} else {
// 			return todoItem
// 		}
// 	}
// }

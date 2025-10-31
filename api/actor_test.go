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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store.OpenSession(ctx, dataFile)
	store.StartActor(ctx)

	// startup the api actor to open the channels
	go func() {
		Actor()
	}()
	m.Run()
	testDataResults := "C:\\Users\\piers.kenyon\\AppData\\Local\\appcli\\testsresults.json"
	store.SaveSession(ctx, testDataResults)
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
	dummyTasks := internal.GenerateDummyTasks(1)
	dummyTask := dummyTasks[0]

	t.Parallel()

	var wg sync.WaitGroup
	for i := range numClients {
		wg.Add(1)
		go func(clientId int) {
			defer wg.Done()
			dummy := dummyTask
			dummy.Description = fmt.Sprintf("%d %s", clientId, dummy.Description)
			jsonData, _ := encodeJsonBodyItem(dummy)
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
			// t.Logf("clientID [%d] created: [%d] [%s]\n", clientId, createdTodo.Line, createdTodo.Description)
			// t.Logf("clientID [%d] created: [%s] expected [%s]\n", clientId, createdTodo.Description, dummy.Description)
			if createdTodo.Description != dummy.Description {
				t.Errorf("create failed, unmatched descriptions, created: [%s] expected [%s]\n", createdTodo.Description, dummy.Description)
			}
		}(i)
	}
	// wait for all client runs to end
	wg.Wait()
}

// update the same line numbers 1..100 with different runnning client ids
func TestUpdate(t *testing.T) {
	var tests = []int64{
		1761837306757785900, 1761837307002502800, 1761837307237967100, 1761837307466795800, 1761837307696710300,
		1761837307919608500, 1761837308134837300, 1761837308377729900, 1761837308603689600, 1761837308840513600,
	}
	dummyTasks := internal.GenerateDummyTasks(1)
	dummyTask := dummyTasks[0]
	lineNumbers := len(tests)
	numClients := 10
	apiCreate := "/update"

	t.Parallel()

	var wg sync.WaitGroup
	// lineNumber is the line item we are updating
	for i := range numClients {
		wg.Add(1)
		go func(clientId int) {
			defer wg.Done()
			for j := range lineNumbers {
				dummy := dummyTask
				dummy.Line = tests[j]
				jsonData, _ := encodeJsonBodyItem(dummy)
				payload := bytes.NewBuffer(jsonData)
				req := httptest.NewRequest(http.MethodPut, apiCreate, payload)
				w := httptest.NewRecorder()
				t.Logf("running update task for Client:%d, %d", clientId, dummy.Line)
				UpdateTask(w, req)

				if w.Result().StatusCode != http.StatusOK {
					t.Errorf("client id %d %d, wanted: %d got:%d", clientId, dummy.Line, http.StatusOK, w.Result().StatusCode)
				} else if w.Body == nil {
					t.Error("update failed, bad body returned")
				}
				updatedTask := decodeJsonRecorderBodyItem(w.Body.Bytes())
				// assert fields here
				if updatedTask.Description != dummy.Description {
					t.Error("update failed, expected descriptions to match")
				}
			}
		}(i)
	}
	// wait for all client runs to end
	wg.Wait()
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

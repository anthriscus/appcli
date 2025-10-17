package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	dataStorageFolderName string      = "appcli"
	dataFileName          string      = "todolist.json"
	errorFileName         string      = "appcliError.log"
	activityFileName      string      = "appcliActivity.log"
	openFlag              int         = os.O_RDWR | os.O_CREATE
	openTruncateFlag      int         = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	readwriteFileMode     os.FileMode = 0600
	traceIdKey            contextKey  = "TraceID"
)

// context key type
type contextKey string

// holds our item
type TodoListItem struct {
	Line        int       `json:"line"` // tags just to show understanding of useage for flipping case in the file.
	Description string    `json:"description"`
	State       int       `json:"state"`
	Created     time.Time `json:"created"`
	Id          string    `json:"id"`
}

type TodoListItems map[int]TodoListItem

// "not started", "started", or "completed", "other etc"
// set them up as consts
const (
	StateNotStarted int = iota
	StateStarted
	StateCompleted
)

// enum equivalent string of status
var statusName = map[int]string{
	StateNotStarted: "Not started",
	StateStarted:    "Started",
	StateCompleted:  "Completed",
}

var errorLogger appLogger
var activityLogger appLogger

func main() {
	// input flags
	var flagAdd = flag.String("add", "", "add todolist item (\"description\")")
	var flagUpdate = flag.Int("update", 0, "update task item description (id -description \"new description\")")
	var flagNotStart = flag.Int("notstart", 0, "set task item id number to not started ( id )")
	var flagStart = flag.Int("start", 0, "start a task item ( id )")
	var flagComplete = flag.Int("complete", 0, "complete a task item id ( id )")
	var flagDelete = flag.Int("delete", 0, "delete a task item id number ( id )")
	var flagList = flag.Bool("list", false, "list items in the todolist item ( with additional optional -taskid num to show one item)")

	// additional flag required for description updates
	var taskDescription string
	flag.Func("description", "use this with -update for the update description text -description \"new text\"", func(s string) error {
		if len(s) == 0 {
			return errors.New("value of description needs to be supplied")
		} else {
			taskDescription = s
		}
		return nil
	})
	// additional flag for list filter
	var taskId int
	flag.Func("taskid", "optional, use this -taskid with -list for one task", func(s string) error {
		if i, ok := strconv.Atoi(s); ok != nil {
			return errors.New("value of taskid needs to be supplied")
		} else {
			taskId = i
		}
		return nil
	})

	// but TODO "github.com/google/uuid" will provide a better one
	id := generateId()
	ctx := context.WithValue(context.Background(), traceIdKey, id)

	// resolve the appdata data sub folder
	dir, err := createAppDataFolder(dataStorageFolderName)
	if err != nil {
		// don't have a file logger yet!
		fmt.Printf("Error:%s", "Cannot establish working data folder")
		return
	}

	// wire up loggers
	errorLogName := dir + "\\" + errorFileName
	if errorFile, err := openLogFile(errorLogName); err == nil {
		defer errorFile.Close()
		errorLogoptions := errorOptions()
		errorLogger.log = setupLogger(errorFile, errorLogoptions)
	}
	activityLogName := dir + "\\" + activityFileName
	if activityFile, err := openLogFile(activityLogName); err == nil {
		defer activityFile.Close()
		activityLogoptions := activityOptions()
		activityLogger.log = setupLogger(activityFile, activityLogoptions)
	}

	// init / pickup current list before process command
	storageFile := fmt.Sprintf("%s\\%s", dir, dataFileName)
	todoList, _ := restore(ctx, storageFile)

	// // grab the flag input state from command line
	flag.Parse()

	// process the flags
	switch {
	case *flagAdd != "":
		nextKey := addTask(ctx, todoList, *flagAdd)
		listTask(todoList, nextKey)
	case *flagUpdate > 0 && len(taskDescription) > 0:
		descriptionChange(ctx, todoList, *flagUpdate, taskDescription)
		listTask(todoList, *flagUpdate)
	case *flagNotStart > 0:
		stateChange(ctx, todoList, *flagNotStart, StateNotStarted)
		listTask(todoList, *flagNotStart)
	case *flagStart > 0:
		stateChange(ctx, todoList, *flagStart, StateStarted)
		listTask(todoList, *flagStart)
	case *flagComplete > 0:
		stateChange(ctx, todoList, *flagComplete, StateCompleted)
		listTask(todoList, *flagComplete)
	case *flagDelete > 0:
		deleteTask(ctx, todoList, *flagDelete)
		listTask(todoList, -1)
	case *flagList:
		listTask(todoList, taskId)
	}

	// write back to the file
	save(ctx, storageFile, todoList)
}

// todolist item constructor
func newTodoListItem(description string, state int, line int) TodoListItem {
	item := TodoListItem{
		Id:          generateId(),
		Description: description,
		State:       state,
		Created:     time.Now().UTC(),
		Line:        line,
	}
	return item
}

// save list back to json
func save(ctx context.Context, storageFile string, list TodoListItems) error {

	if data, err := json.Marshal(list); err != nil {
		errorLogger.log.ErrorContext(ctx, "Save failed converting todo list to json", "err", err)
		return err
	} else {
		if destination, err := os.OpenFile(storageFile, openTruncateFlag, readwriteFileMode); err != nil {
			errorLogger.log.ErrorContext(ctx, "Save failed getting file", "err", err, "storageFile", storageFile)
			return err
		} else {
			defer destination.Close()
			if _, err := destination.Write(data); err != nil {
				errorLogger.log.ErrorContext(ctx, "Save to file failed ", "err", err, "storageFile", storageFile)
				return err
			}
		}
	}
	activityLogger.log.InfoContext(ctx, "Saved data", "storageFile", storageFile)
	return nil
}

// restore from json
func restore(ctx context.Context, storageFile string) (TodoListItems, error) {
	destination, err := os.OpenFile(storageFile, openFlag, readwriteFileMode)
	if err != nil {
		errorLogger.log.ErrorContext(ctx, "Error restoring list file", "err", err, "storageFile", storageFile)
		return TodoListItems{}, err
	}
	if destination != nil {
		defer destination.Close()
	}
	return restoreList(ctx, destination)
}

func restoreList(ctx context.Context, destination io.Reader) (TodoListItems, error) {
	if restored, err := io.ReadAll(destination); err != nil {
		fmt.Println(err)
		fmt.Printf("error restoring data")
		errorLogger.log.ErrorContext(ctx, "Error restoring data", "err", err)
		return TodoListItems{}, err
	} else if len(restored) == 0 {
		// not neccessarily an error
		fmt.Printf("returning empty list restored empty")
		return TodoListItems{}, nil
	} else {
		data := []byte(string(restored))
		restoredList := TodoListItems{}
		err := json.Unmarshal(data, &restoredList)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("returning empty list json error")
			errorLogger.log.ErrorContext(ctx, "Error restoring list from json", "err", err)
			return TodoListItems{}, nil
		}
		return restoredList, nil
	}
}

// more unique (ish) id perhaps
// todo consider using https://github.com/google/uuid
func generateId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// task list report
func listTask(list TodoListItems, index int) {
	fmt.Printf("List length:%d\n", len(list))

	listTaskHeader()
	if len(list) > 0 {
		if record, ok := list[index]; ok {
			listTaskLine(record)
		} else {
			itemKeys := collectKeys(list)
			slices.Sort(itemKeys)
			for _, i := range itemKeys {
				listTaskLine(list[i])
			}
		}
	}
}

func listTaskHeader() {
	fmt.Printf("%s\t%s\t\t%s\n", "ID", "Status", "Description")
	fmt.Printf("%s\t%s\t%s\n", strings.Repeat("-", 1), strings.Repeat("-", 12), strings.Repeat("-", 120))
}

func listTaskLine(listItem TodoListItem) {
	fmt.Printf("%d\t%s\t%s\t[%s]\n", listItem.Line, statusName[listItem.State], listItem.Description, listItem.Created.Format(time.RFC822))
}

// delete a task
func deleteTask(ctx context.Context, list TodoListItems, index int) {
	if len(list) > 0 {
		if record, ok := list[index]; ok {
			fmt.Printf("Deleting item: %d\n", index)
			before := record.Description
			delete(list, index)
			activityLogger.log.InfoContext(ctx, "Deleted item", "ID", index, "before", before)
		}
	}
}

// change the state
func stateChange(ctx context.Context, list TodoListItems, index int, state int) {
	if len(list) > 0 {
		if record, ok := list[index]; ok {
			fmt.Printf("Current state: %s", statusName[list[index].State])
			fmt.Printf("Changing task %d state to : %s\n", index, statusName[state])
			before := statusName[list[index].State]
			after := statusName[state]
			record.State = state
			list[index] = record
			activityLogger.log.InfoContext(ctx, "Updated item status", "ID", index, "before", before, "after", after)
		}
	}
}

func addTask(ctx context.Context, list TodoListItems, newItem string) int {
	itemKeys := collectKeys(list)
	nextKey := highestKey(itemKeys) + 1
	item := newTodoListItem(newItem, StateNotStarted, nextKey)
	list[nextKey] = item
	activityLogger.log.InfoContext(ctx, "Added item", "ID", nextKey, "descriptiont", newItem)
	return nextKey
}

func descriptionChange(ctx context.Context, list TodoListItems, index int, newDescription string) {
	if len(list) > 0 {
		if record, ok := list[index]; ok {
			fmt.Printf("Current description: %s", list[index].Description)
			fmt.Printf("Changing task %d description to : %s\n", index, newDescription)
			before := record.Description
			record.Description = newDescription
			list[index] = record
			activityLogger.log.InfoContext(ctx, "Updated item description", "ID", index, "before", before, "after", newDescription)
		}
	}
}

// fetch the number keys from the map
func collectKeys(data TodoListItems) []int {
	keys := make([]int, 0, len(data))
	for i := range data {
		keys = append(keys, i)
	}
	return keys
}

func highestKey(keys []int) int {
	key := 1
	for _, i := range keys {
		if i > key {
			key = i
		}
	}
	return key
}

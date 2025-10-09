package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	dataFileName       string      = "todolist.json"
	openmode           int         = os.O_RDWR | os.O_CREATE
	opentruncatemode   int         = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	readwrite          os.FileMode = 0600
	friendlyDateFormat string      = time.RFC3339
)

// holds our item
type TodoListItem struct {
	Line        int    `json:"line"`
	Description string `json:"description"`
	State       int    `json:"state"`
	Created     string `json:"created"` // friendly date
	Id          string `json:"id"`
}

// "not started", "started", or "completed", "other etc"
// set them up as consts
const (
	StateNotStarted int = iota
	StateStarted
	StateCompleted
)

// enum string them
var statusName = map[int]string{
	StateNotStarted: "not started",
	StateStarted:    "started",
	StateCompleted:  "completed",
}

func main() {
	// not sure about var naming for flags here..
	var flagTaskStatus = flag.Int("status", 0, "set the task status")
	var flagAdd = flag.String("add", "", "add to todolist item")
	var flagUpdate = flag.Int("update", -1, "update to todolist item")
	var flagDelete = flag.Int("delete", -1, "delete item in todolist item")
	var flagList = flag.Int("list", -1, "List items in to todolist item (0 for all)")
	// var flagTask = flag.String("name", "", "task name")

	// scratch debug info
	fmt.Printf("program arguments length: %d\n", len(os.Args))
	fmt.Printf("program arguments: %s", os.Args)
	fmt.Printf("\n")
	fmt.Printf("lastArgument is: %s\n", lastArgument(os.Args))
	fmt.Printf("appname is %s\n", appName())

	// for loop understanding
	fmt.Println("")
	if len(os.Args) > 1 {
		for index := range len(os.Args[1:]) {
			fmt.Printf("arguments item: %s\n", os.Args[index+1])
		}
	}

	// our global list
	todoList := []TodoListItem{}

	// file resolver stuff for filer.go ?
	storageDir, storageErr := storagePath()
	if storageErr == nil {
		fmt.Printf("Storage path:  %s\n", storageDir)
	} else {
		fmt.Printf("Storage path check failed:  %s\n", storageErr)
	}

	storageFile := fmt.Sprintf("%s\\%s", storageDir, dataFileName)
	fmt.Printf("Will be saving data to:  %s\n", storageFile)
	// end file resolver

	// pickup current list before process command
	todoList = restore(storageFile)

	fmt.Printf("BeforeActions: Items in todoList: %d\n", len(todoList))
	fmt.Println(todoList)
	fmt.Println(strings.Repeat("*", 25))

	// grab the flag input state from command line
	flag.Parse()

	fmt.Printf("Status is: %d\n", *flagTaskStatus)
	if *flagTaskStatus > -1 {
		fmt.Printf("Status flag is: %d\n", *flagTaskStatus)
		fmt.Printf("status const is %d\n", StateStarted)
		fmt.Printf("status string is %s\n", statusName[StateStarted])
		fmt.Println(statusName)
		fmt.Printf("statusName length: %d\n", len(statusName))
		var flag = *flagTaskStatus
		checkStatus(flag)
	}

	if *flagAdd != "" {
		fmt.Printf("%s\n", "wanting to add a todo item")
		fmt.Printf("[%s]\n", *flagAdd)
		item := newTodoListItem(*flagAdd, StateNotStarted, len(todoList)+1)
		todoList = append(todoList, *item)
	}

	if *flagUpdate != -1 {
		// todo
	}
	if *flagDelete != -1 {
		fmt.Printf("FlagDelete [%d]\n", *flagDelete)
		todoList = deleteTask(todoList, *flagDelete)
		listTask(todoList, -1)
	}
	if *flagList > -1 {
		fmt.Printf("FlagList [%d]\n", *flagList)
		listTask(todoList, *flagList)
	}

	// write to the file
	saveList := []byte(stringifyList(todoList))
	save(storageFile, saveList)
}

func appName() string {
	return strings.Split(filepath.Base(os.Args[0]), ".")[0]
}

func checkStatus(status int) {
	if status >= 0 && status <= len(statusName) {
		fmt.Printf("Status is: %s\n", statusName[status])
	} else {
		panic(fmt.Errorf("unknown state: %d", status))
	}
}

func storagePath() (string, error) {
	storageDir, storageErr := os.UserCacheDir()
	storageDir = createAppPath(fmt.Sprintf("%s\\%s", storageDir, appName()))
	return storageDir, storageErr
}

func createAppPath(fileFolder string) string {
	os.Mkdir(fileFolder, readwrite)
	return fileFolder
}

func createDataFile(fileName string, mode int) (*os.File, error) {
	fi, err := os.OpenFile(fileName, mode, readwrite)
	return fi, err
}

func stringifyList(list []TodoListItem) string {
	// back as byte
	data, _ := json.Marshal(list)
	// then string it
	return string(data)
}

func lastArgument(args []string) string {
	if len(args) > 1 {
		return args[len(args)-1]
	}
	return ""
}

func newTodoListItem(description string, state int, line int) *TodoListItem {
	item := TodoListItem{
		Id:          generateId(),
		Description: description,
		State:       state,
		Created:     generatedFriendlyDate(),
		Line:        line,
	}
	return &item
}

func save(storageFile string, data []byte) {
	destination, err := createDataFile(storageFile, opentruncatemode)
	if destination != nil {
		defer destination.Close()
	}
	if err == nil {
		saveData(destination, data)
	}
}

func saveData(w io.Writer, data []byte) int {
	bytes, _ := w.Write(data)
	return bytes
}

func restore(storageFile string) []TodoListItem {
	destination, err := createDataFile(storageFile, openmode)
	if destination != nil {
		defer destination.Close()
	}
	if err != nil {
		return []TodoListItem{}
	}
	return restoreList(destination)
}

func restoreList(destination io.Reader) []TodoListItem {
	restored := restoreData(destination)
	if len(restored) == 0 {
		fmt.Printf("returning empty list restored empty")
		return []TodoListItem{}
	}
	data := []byte(string(restored))
	restoredList := []TodoListItem{}
	jsonError := json.Unmarshal(data, &restoredList)
	if jsonError != nil {
		fmt.Println(jsonError)
		fmt.Printf("returning empty list json error")
		return []TodoListItem{}
	}
	return restoredList
}

func restoreData(r io.Reader) []byte {
	var restored, _ = io.ReadAll(r)
	return restored
}

// could use other choices of id layout/format here. guid etc for safer uniqueness?
func generateId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// user friendly date for reports etc
func generatedFriendlyDate() string {
	t := time.Now().UTC()
	return fmt.Sprintf("%s", t.Format(friendlyDateFormat))
}

func listTask(l []TodoListItem, index int) {
	fmt.Printf("%s Todo List items %s\n", strings.Repeat("*", 3), strings.Repeat("*", 3))
	if len(l) > 0 {
		if index > 0 && index <= len(l)+1 {
			fmt.Printf("%+v\n", l[index-1])
		} else {
			for i := range len(l) {
				fmt.Printf("%+v\n", l[i])
			}
		}
	}
}

func deleteTask(l []TodoListItem, index int) []TodoListItem {
	if len(l) > 0 {
		if index > 0 && index <= len(l) {
			fmt.Printf("%s Deleting item %d %s\n", strings.Repeat("*", 3), index, strings.Repeat("*", 3))
			l = slices.Delete(l, index-1, index)
		}
	}
	return l
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	dataFileName       string      = "todolist.json"
	openmode           int         = os.O_RDWR | os.O_CREATE
	readwrite          os.FileMode = 0600
	friendlyDateFormat string      = time.RFC3339
)

// holds our item
type TodoListItem struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	State       int    `json:"state"`
	Created     string `json:"created"` // friendly date
	Line        int    `json:"line"`
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
	var flagList = flag.String("list", "", "List items in to todolist item")
	// var flagTask = flag.String("name", "", "task name")

	// scratch debug info
	fmt.Printf("program arguments length: %d\n", len(os.Args))
	fmt.Printf("program arguments: %s", os.Args)
	fmt.Printf("\n")
	fmt.Printf("lastArgument is: %s\n", lastArgument(os.Args))
	fmt.Printf("appname is %s\n", appName())

	// our global list
	todoList := []TodoListItem{}

	// grab the flag input state
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

		// try a for to see effect of many
		for i := range 5 {
			item := newTodoListItem(*flagAdd, StateNotStarted, len(todoList)+1)
			todoList = append(todoList, *item)
			i++
		}
	}
	if *flagUpdate != -1 {
		// todo
	}
	if *flagDelete != -1 {
		// todo
	}
	if *flagList != "" {
		//todo
	}

	// for loop understanding
	fmt.Println("")
	if len(os.Args) > 1 {
		for index := range len(os.Args[1:]) {
			fmt.Printf("arguments item: %s\n", os.Args[index+1])
		}
	}

	// for loop understanding
	// fmt.Println("line report etc")
	// for index := range len(todolist) {
	// 	fmt.Printf("arguments item: %s\n", os.Args[index+1])
	// }

	// file stuff filer.go ?
	storageDir, storageErr := storagePath()
	if storageErr == nil {
		fmt.Printf("Storage path:  %s\n", storageDir)
	} else {
		fmt.Printf("Storage path check failed:  %s\n", storageErr)
	}

	storageFile := fmt.Sprintf("%s\\%s", storageDir, dataFileName)
	fmt.Printf("Will saving data to:  %s\n", storageFile)

	destination, err := createDataFile(storageFile)
	if err == nil {
		fmt.Printf("Should have created:  %s\n", storageFile)
	}

	// debug info
	fmt.Printf("os file pointer is %p\n", destination)

	// debug data in the todo list
	s := stringifyList(todoList)
	fmt.Printf("Data list as json:\n%s\n", s)

	// write to the file
	// saveData := []byte(stringifyList(todoList))
	// save(destination, saveData)

	// try a restore
	// restored := restore(destination)
	// fmt.Printf("restored string \n%s\n", string(restored))
	// data := []byte(string(restored))
	// restoredList := []TodoListItem{}
	// restoredError := json.Unmarshal(data, &restoredList)
	// fmt.Println(restoredError)
	// fmt.Println(restoredList)

	checkRestore(destination)
}

// // check the read
// var restored, _ = io.ReadAll(destination)
// fmt.Printf("restored string \n%s\n", string(restored))
// //

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

func createDataFile(fileName string) (*os.File, error) {
	fi, err := os.OpenFile(fileName, openmode, readwrite)
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

func save(w io.Writer, data []byte) int {
	bytes, _ := w.Write(data)
	return bytes
}

func restore(r io.Reader) []byte {
	var restored, _ = io.ReadAll(r)
	return restored
}

func checkRestore(destination io.Reader) {
	// try a restore and look at contents
	restored := restore(destination)
	fmt.Printf("restored string \n%s\n", string(restored))
	data := []byte(string(restored))
	restoredList := []TodoListItem{}
	restoredError := json.Unmarshal(data, &restoredList)
	fmt.Println(restoredError)
	fmt.Println(restoredList)
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

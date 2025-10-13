package main

import (
	"encoding/json"
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
	openFlag              int         = os.O_RDWR | os.O_CREATE
	openTruncateFlag      int         = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	readwriteFileMode     os.FileMode = 0600
	friendlyDateFormat    string      = time.RFC3339
)

// holds our item
type TodoListItem struct {
	Line        int    `json:"line"`
	Description string `json:"description"`
	State       int    `json:"state"`
	Created     string `json:"created"` // friendly date
	Id          string `json:"id"`
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

func main() {
	// input flags
	var flagAdd = flag.String("add", "", "add todolist item (\"description\")")
	var flagUpdate = flag.String("update", "", "update task item description (\"description\" id)")
	var flagNotStart = flag.Bool("notstart", false, "set task item id number to not started ( id )")
	var flagStart = flag.Bool("start", false, "start a task item ( id )")
	var flagComplete = flag.Bool("complete", false, "complete a task item id ( id )")
	var flagDelete = flag.Int("delete", -1, "delete a task item id number ( id )")
	var flagList = flag.Bool("list", false, "list items in the todolist item ( id )")

	// resolve the appdata data sub folder
	dir := createAppDataFolder(dataStorageFolderName)
	if len(dir) == 0 {
		fmt.Printf("Error:%s", "Cannot establish working data folder")
		return
	}

	// init / pickup current list before process command
	storageFile := fmt.Sprintf("%s\\%s", dir, dataFileName)
	todoList := restore(storageFile)

	// // grab the flag input state from command line
	flag.Parse()

	// process the flags
	switch {
	case *flagAdd != "":
		itemKeys := collectKeys(todoList)
		nextKey := highestKey(itemKeys) + 1
		item := newTodoListItem(*flagAdd, StateNotStarted, nextKey)
		todoList[nextKey] = *item
		listTask(todoList, nextKey)
	case *flagUpdate != "":
		index := argumentsFlagIndex(true, os.Args)
		descriptionChange(todoList, index, *flagUpdate)
		listTask(todoList, index)
	case *flagNotStart:
		index := argumentsFlagIndex(*flagNotStart, os.Args)
		stateChange(todoList, index, StateNotStarted)
		listTask(todoList, index)
	case *flagStart:
		index := argumentsFlagIndex(*flagStart, os.Args)
		stateChange(todoList, index, StateStarted)
		listTask(todoList, index)
	case *flagComplete:
		index := argumentsFlagIndex(*flagComplete, os.Args)
		stateChange(todoList, index, StateCompleted)
		listTask(todoList, index)
	case *flagDelete != -1:
		fmt.Printf("FlagDelete [%d]\n", *flagDelete)
		deleteTask(todoList, *flagDelete)
		listTask(todoList, -1)
	case *flagList:
		index := argumentsFlagIndex(*flagList, os.Args)
		listTask(todoList, index)
	}

	// // write back to the file
	saveList := []byte(stringifyList(todoList))
	save(storageFile, saveList)
}

// find the command line taskitem number
func argumentsFlagIndex(flag bool, args []string) int {
	if flag && len(args) >= 3 {
		index, err := strconv.Atoi(args[len(args)-1])
		if err == nil {
			return index
		}
	}
	return 0
}

func createAppDataFolder(applicationName string) string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	dir = dir + "\\" + applicationName
	err = os.MkdirAll(dir, readwriteFileMode)
	if err != nil {
		return ""
	}
	return dir
}

// todolist item constructor
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

func stringifyList(list TodoListItems) string {
	// back as byte
	data, _ := json.Marshal(list)
	// then string it
	return string(data)
}

// save list back to json
func save(storageFile string, data []byte) {
	destination, err := os.OpenFile(storageFile, openTruncateFlag, readwriteFileMode)
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

// restore from json
func restore(storageFile string) TodoListItems {
	destination, err := os.OpenFile(storageFile, openFlag, readwriteFileMode)
	if err != nil {
		return TodoListItems{}
	}
	if destination != nil {
		defer destination.Close()
	}
	return restoreList(destination)
}

func restoreList(destination io.Reader) TodoListItems {
	restored := restoreData(destination)
	if len(restored) == 0 {
		fmt.Printf("returning empty list restored empty")
		return TodoListItems{}
	}
	data := []byte(string(restored))
	restoredList := TodoListItems{}
	jsonError := json.Unmarshal(data, &restoredList)
	if jsonError != nil {
		fmt.Println(jsonError)
		fmt.Printf("returning empty list json error")
		return TodoListItems{}
	}
	return restoredList
}

func restoreData(r io.Reader) []byte {
	var restored, _ = io.ReadAll(r)
	return restored
}

// more unique (ish) id perhaps
func generateId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// date string for task item
func generatedFriendlyDate() string {
	t := time.Now().UTC()
	return "%s" + t.Format(friendlyDateFormat)
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
	fmt.Printf("%d\t%s\t%s\n", listItem.Line, statusName[listItem.State], listItem.Description)
}

// delete a task
func deleteTask(list TodoListItems, index int) {
	if len(list) > 0 {
		if _, ok := list[index]; ok {
			fmt.Printf("Deleting item: %d\n", index)
			delete(list, index)
		}
	}
}

// change the state
func stateChange(list TodoListItems, index int, state int) {
	if len(list) > 0 {
		if record, ok := list[index]; ok {
			fmt.Printf("Current state: %s", statusName[list[index].State])
			fmt.Printf("Changing task %d state to : %s\n", index, statusName[state])
			record.State = state
			list[index] = record
		}
	}
}

func descriptionChange(list TodoListItems, index int, newDescription string) {
	if len(list) > 0 {
		if record, ok := list[index]; ok {
			fmt.Printf("Current description: %s", list[index].Description)
			fmt.Printf("Changing task %d description to : %s\n", index, newDescription)
			record.Description = newDescription
			list[index] = record
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

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
	dataStorageFolder  string      = "appcli"
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
	var flagAdd = flag.String("add", "", "add todolist item")
	var flagUpdate = flag.String("update", "", "update task item id description")
	var flagNotStart = flag.Bool("notstart", false, "not start a task item id number (not started)")
	var flagStart = flag.Bool("start", false, "start a task item id number (starte)")
	var flagComplete = flag.Bool("complete", false, "complete a task item id number (completed)")
	var flagDelete = flag.Int("delete", -1, "delete a task item id number")
	var flagList = flag.Bool("list", false, "list items in the todolist item")

	// resolver the data folder
	storageDir, storageErr := storagePath()
	if storageErr != nil {
		fmt.Printf("Storage path check failed:  %s\n", storageErr)
		return
	}
	storageFile := fmt.Sprintf("%s\\%s", storageDir, dataFileName)

	// init / pickup current list before process command
	todoList := restore(storageFile)

	// // grab the flag input state from command line
	flag.Parse()

	// process the flags
	if *flagAdd != "" {
		itemKeys := collectKeys(todoList)
		nextKey := highestKey(itemKeys) + 1
		item := newTodoListItem(*flagAdd, StateNotStarted, nextKey)
		todoList[nextKey] = *item
	}
	if *flagUpdate != "" {
		index := argumentsFlagIndex(true, os.Args)
		descriptionChange(todoList, index, *flagUpdate)
	}
	if *flagNotStart {
		index := argumentsFlagIndex(*flagNotStart, os.Args)
		stateChange(todoList, index, StateNotStarted)
		listTask(todoList, index)
	}
	if *flagStart {
		index := argumentsFlagIndex(*flagStart, os.Args)
		stateChange(todoList, index, StateStarted)
		listTask(todoList, index)
	}
	if *flagComplete {
		index := argumentsFlagIndex(*flagComplete, os.Args)
		stateChange(todoList, index, StateCompleted)
		listTask(todoList, index)
	}
	if *flagDelete != -1 {
		fmt.Printf("FlagDelete [%d]\n", *flagDelete)
		deleteTask(todoList, *flagDelete)
		listTask(todoList, -1)
	}
	if *flagList {
		fmt.Printf("FlagList [%t]\n", *flagList)
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

// for sub folder
func appName() string {
	// return strings.Split(filepath.Base(os.Args[0]), ".")[0] // found problem with debugger renaming exe name!
	return dataStorageFolder
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

// restore from json
func restore(storageFile string) TodoListItems {
	destination, err := createDataFile(storageFile, openmode)
	if destination != nil {
		defer destination.Close()
	}
	if err != nil {
		return TodoListItems{}
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
	return fmt.Sprintf("%s", t.Format(friendlyDateFormat))
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
			fmt.Printf("Changing task %d state to : %s\n", index, statusName[state])
			fmt.Printf("Current state: %s", statusName[list[index].State])
			record.State = state
			list[index] = record
		}
	}
}

func descriptionChange(list TodoListItems, index int, newDescription string) {
	if len(list) > 0 {
		if record, ok := list[index]; ok {
			fmt.Printf("Changing task %d description to : %s\n", index, newDescription)
			fmt.Printf("Current description: %s", list[index].Description)
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

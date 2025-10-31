package store

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/anthriscus/appcli/logging"
)

const (
	StateNotStarted int = iota
	StateStarted
	StateCompleted
)

// enum equivalent string of status
var StatusName = map[int]string{
	StateNotStarted: "Not started",
	StateStarted:    "Started",
	StateCompleted:  "Completed",
}

// note point on unique keys in
// https://go.dev/ref/spec#Composite_literals
// https://go.dev/ref/spec#Order_of_evaluation

type TodoListItem struct {
	Line        int64     `json:"line"` // tags just to show understanding of useage for flipping case in the file.
	Description string    `json:"description"`
	State       int       `json:"state"`
	Created     time.Time `json:"created"`
	Id          int64     `json:"id"`
}

type TodoListItems map[int64]TodoListItem

var storeActor *StoreChannels

// more unique (ish) id perhaps
// todo consider using https://github.com/google/uuid
// note point on unique keys in
// https://go.dev/ref/spec#Composite_literals
// https://go.dev/ref/spec#Order_of_evaluation

func GenerateId() int64 {
	// return fmt.Sprintf("%d", time.Now().UnixNano())
	// need uuid not a int64 but that means refactor of where we use the line/id
	// added a rand otherwise multi runners will get the same int64 id
	// but all this is just a demo/mock app to learn stuff. sometimes our under load testing is generating same key for items.
	t := time.Now().UnixNano() + rand.Int63n(10000)
	atomic.AddInt64(&t, 1)
	return t
}

func StartActor(ctx context.Context) {
	storeActor = NewStoreChannels(ctx)
}

func AddTask(ctx context.Context, newItem string) (int64, error) {

	if !isDescription(newItem) {
		return 0, errors.New("description cannot be empty")
	}
	// itemKeys := collectKeys(sessionDatabase)
	// nextKey := highestKey(itemKeys) + 1
	// cannot use next key because multiple go routines were ending up with same line no/next key.
	// just using the psuedo random int64 number for now. to avoid changing all the code for ids to uuid at this point.
	// So needs refactor !
	item := newTodoListItem(newItem, StateNotStarted)
	// sessionDatabase[nextKey] = item
	record := TodoListRecord{item: item}
	storeActor.Write(record)

	logging.Log().InfoContext(ctx, "Added item", "ID", item.Id, "description", newItem)
	return item.Id, nil
}

func newTodoListItem(description string, state int) TodoListItem {
	newId := GenerateId()
	item := TodoListItem{
		Id:          newId,
		Description: description,
		State:       state,
		Created:     time.Now().UTC(),
		Line:        newId,
	}
	return item
}

func DescriptionChange(ctx context.Context, index int64, newDescription string) error {
	if len(sessionDatabase) > 0 {
		if !isDescription(newDescription) {
			return errors.New("description cannot be empty")
		} else if record, ok := sessionDatabase[index]; !ok {
			return fmt.Errorf("cannot find item %d", index)
		} else {
			fmt.Printf("Current description: %s\n", sessionDatabase[index].Description)
			fmt.Printf("Changing task %d description to : %s\n", index, newDescription)
			before := record.Description
			record.Description = newDescription
			sessionDatabase[index] = record
			logging.Log().InfoContext(ctx, "Updated item description", "ID", index, "before", before, "after", newDescription)
			return nil
		}
	}
	return errors.New("cannot set description")
}

// change the state
func StateChange(ctx context.Context, index int64, state int) error {
	if len(sessionDatabase) > 0 {
		if !isState(state) {
			return errors.New("state is out of range")
		} else if record, ok := sessionDatabase[index]; !ok {
			return fmt.Errorf("cannot find item %d", index)
		} else {
			fmt.Printf("Current state: %s\n", StatusName[sessionDatabase[index].State])
			fmt.Printf("Changing task %d state to : %s\n", index, StatusName[state])
			before := StatusName[sessionDatabase[index].State]
			after := StatusName[state]
			fmt.Printf("before:%s after:%s\n", before, after)
			record.State = state
			sessionDatabase[index] = record
			logging.Log().InfoContext(ctx, "Updated item status", "ID", index, "before", before, "after", after)
			return nil
		}
	}
	return errors.New("cannote set state")
}

func UpdateTask(ctx context.Context, item TodoListItem) (TodoListItem, error) {
	if !isDescription(item.Description) {
		return TodoListItem{}, errors.New("description cannot be empty")
	} else if !isState(item.State) {
		return TodoListItem{}, errors.New("state is out of range")
	}
	//if current, ok := sessionDatabase[item.Line]; !ok {
	record := storeActor.Read(item.Line)
	ok := record.ok
	current := record.item
	if !ok {
		empty := TodoListItem{}
		return empty, fmt.Errorf("error: %s", "item not found")
	} else {
		// only update the task and description
		current.Description = item.Description
		current.State = item.State
		// sessionDatabase[item.Line] = current
		record := TodoListRecord{item: current}
		storeActor.Write(record)
		index := item.Line
		after := item.Description
		logging.Log().InfoContext(ctx, "Updated item", "ID", index, "description", after)
		return current, nil
	}
}

// delete a task
func DeleteTask(ctx context.Context, index int64) error {
	if len(sessionDatabase) > 0 {
		if record, ok := sessionDatabase[index]; !ok {
			return errors.New("item not found")
		} else {
			fmt.Printf("Deleting item: %d\n", index)
			before := record.Description
			fmt.Printf("before:%s\n", before)
			delete(sessionDatabase, index)
			logging.Log().InfoContext(ctx, "Deleted item", "ID", index, "before", before)
			return nil
		}
	}
	return errors.New("item cannot be deleted")
}

// task list report
func ListTask(index int64) {
	fmt.Printf("\nList length:%d\n", len(sessionDatabase))

	listTaskHeader()
	if len(sessionDatabase) > 0 {
		if record, ok := sessionDatabase[index]; ok {
			listTaskLine(record)
		} else {
			itemKeys := collectKeys(sessionDatabase)
			slices.Sort(itemKeys)
			for _, i := range itemKeys {
				listTaskLine(sessionDatabase[i])
			}
		}
	}
}

func listTaskHeader() {
	fmt.Printf("%s\t%s\t\t%s\n", "ID", "Status", "Description")
	fmt.Printf("%s\t%s\t%s\n", strings.Repeat("-", 1), strings.Repeat("-", 12), strings.Repeat("-", 120))
}

func listTaskLine(listItem TodoListItem) {
	fmt.Printf("%d\t%s\t%s\t[%s]\n", listItem.Line, StatusName[listItem.State], listItem.Description, listItem.Created.Format(time.RFC822))
}

// fetch the number keys from the map
func collectKeys(data TodoListItems) []int64 {
	keys := make([]int64, 0, len(data))
	for i := range data {
		keys = append(keys, i)
	}
	return keys
}

func highestKey(keys []int) int {
	key := 0
	for _, i := range keys {
		if i > key {
			key = i
		}
	}
	return key
}

func isDescription(description string) bool {
	return description != ""
}

func isState(state int) bool {
	states := make([]int, 0, len(StatusName))
	for i := range StatusName {
		states = append(states, i)
	}
	return slices.Contains(states, state)
}

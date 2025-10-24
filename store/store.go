package store

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
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

type TodoListItem struct {
	Line        int       `json:"line"` // tags just to show understanding of useage for flipping case in the file.
	Description string    `json:"description"`
	State       int       `json:"state"`
	Created     time.Time `json:"created"`
	Id          string    `json:"id"`
}

type TodoListItems map[int]TodoListItem

// more unique (ish) id perhaps
// todo consider using https://github.com/google/uuid
func GenerateId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func AddTask(ctx context.Context, newItem string) (int, error) {

	if !isDescription(newItem) {
		return 0, errors.New("description cannot be empty")
	}
	itemKeys := collectKeys(sessionDatabase)
	nextKey := highestKey(itemKeys) + 1
	item := newTodoListItem(newItem, StateNotStarted, nextKey)
	sessionDatabase[nextKey] = item

	logging.Log().InfoContext(ctx, "Added item", "ID", nextKey, "description", newItem)
	return nextKey, nil
}

func newTodoListItem(description string, state int, line int) TodoListItem {
	item := TodoListItem{
		Id:          GenerateId(),
		Description: description,
		State:       state,
		Created:     time.Now().UTC(),
		Line:        line,
	}
	return item
}

func DescriptionChange(ctx context.Context, index int, newDescription string) error {
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
func StateChange(ctx context.Context, index int, state int) error {
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
	if current, ok := sessionDatabase[item.Line]; !ok {
		empty := TodoListItem{}
		return empty, fmt.Errorf("error: %s", "item not found")
	} else {
		// only update the task and description
		current.Description = item.Description
		current.State = item.State
		sessionDatabase[item.Line] = current
		index := item.Line
		after := item.Description
		logging.Log().InfoContext(ctx, "Updated item", "ID", index, "description", after)
		return current, nil
	}
}

// delete a task
func DeleteTask(ctx context.Context, index int) error {
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
func ListTask(index int) {
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
func collectKeys(data TodoListItems) []int {
	keys := make([]int, 0, len(data))
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

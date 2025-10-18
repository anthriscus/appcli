package store

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
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

func AddTask(ctx context.Context, list TodoListItems, newItem string) int {
	itemKeys := collectKeys(list)
	nextKey := highestKey(itemKeys) + 1
	item := newTodoListItem(newItem, StateNotStarted, nextKey)
	list[nextKey] = item

	// ActivityLogger.Log.InfoContext(ctx, "Added item", "ID", nextKey, "descriptiont", newItem)
	return nextKey
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

func DescriptionChange(ctx context.Context, list TodoListItems, index int, newDescription string) {
	if len(list) > 0 {
		if record, ok := list[index]; ok {
			fmt.Printf("Current description: %s\n", list[index].Description)
			fmt.Printf("Changing task %d description to : %s\n", index, newDescription)
			// before := record.Description
			record.Description = newDescription
			list[index] = record
			// ActivityLogger.Log.InfoContext(ctx, "Updated item description", "ID", index, "before", before, "after", newDescription)
		}
	}
}

// change the state
func StateChange(ctx context.Context, list TodoListItems, index int, state int) {
	if len(list) > 0 {
		if record, ok := list[index]; ok {
			fmt.Printf("Current state: %s\n", StatusName[list[index].State])
			fmt.Printf("Changing task %d state to : %s\n", index, StatusName[state])
			before := StatusName[list[index].State]
			after := StatusName[state]
			fmt.Printf("before:%s after:%s\n", before, after)
			record.State = state
			list[index] = record
			// ActivityLogger.Log.InfoContext(ctx, "Updated item status", "ID", index, "before", before, "after", after)
		}
	}
}

// delete a task

func DeleteTask(ctx context.Context, list TodoListItems, index int) {
	if len(list) > 0 {
		if record, ok := list[index]; ok {
			fmt.Printf("Deleting item: %d\n", index)
			before := record.Description
			fmt.Printf("before:%s\n", before)
			delete(list, index)
			// ActivityLogger.Log.InfoContext(ctx, "Deleted item", "ID", index, "before", before)
		}
	}
}

// task list report
func ListTask(list TodoListItems, index int) {
	fmt.Printf("\nList length:%d\n", len(list))

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
	key := 1
	for _, i := range keys {
		if i > key {
			key = i
		}
	}
	return key
}

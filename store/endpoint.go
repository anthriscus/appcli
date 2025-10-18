package store

import (
	// "context"
	// "encoding/json"
	"errors"
	// "flag"
	"fmt"
	// "io"
	// "os"
	"slices"
	// "strconv"
	// "strings"
	"time"
)

var (
	sampleQuotes = []string{
		"The quick brown fox",
		"jumps over the lazy dog",
		"to be or not to be",
		"that is the question",
		"whether tis nobler in the mind",
		"to suffer the outrageous fortune"}
)

// new item
func Create(candidate TodoListItem) (TodoListItem, error) {
	item := TodoListItem{
		Id:          fmt.Sprintf("%d", time.Now().UnixNano()),
		Description: sampleQuotes[0],
		State:       0,
		Created:     time.Now().UTC(),
		Line:        42,
	}
	return item, nil
}

// get by taskid
func GetByIndex(taskId int) (TodoListItem, error) {
	if sessionDatabase == nil {
		var empty TodoListItem
		return empty, fmt.Errorf("error: %s", "datalist not open")
	}
	if item, ok := sessionDatabase[taskId]; !ok {
		empty := TodoListItem{}
		return empty, fmt.Errorf("error: %s", "item not found")
	} else {
		return item, nil
	}
}

// list items
func GetList() (TodoListItems, error) {
	if sessionDatabase == nil {
		var empty TodoListItems
		return empty, fmt.Errorf("error: %s", "datalist not open")
	}
	return sessionDatabase, nil
}

// update item
func UpdateByIndex(taskId int, item TodoListItem) (TodoListItem, error) {
	if item.Description == "" {
		return TodoListItem{}, errors.New("description cannot be empty")
	} else if !isState(item.State) {
		return TodoListItem{}, errors.New("state is out of range")

	} else {
		// nextKey := taskId
		// newItem := TodoListItem{
		// 	Id:          fmt.Sprintf("%d", time.Now().UnixNano()),
		// 	Description: item.Description,
		// 	State:       item.State,
		// 	Created:     time.Now().UTC(),
		// 	Line:        nextKey,
		// }
		// return newItem, nil
		if sessionDatabase == nil {
			var empty TodoListItem
			return empty, fmt.Errorf("error: %s", "datalist not open")
		}
		if current, ok := sessionDatabase[taskId]; !ok {
			empty := TodoListItem{}
			return empty, fmt.Errorf("error: %s", "item not found")
		} else {
			current.Description = item.Description
			current.State = item.State
			sessionDatabase[taskId] = current
			return current, nil
		}
	}
}

func Delete(taskId int) error {
	return nil
}

func isState(state int) bool {
	states := make([]int, 0, len(StatusName))
	for i := range StatusName {
		states = append(states, i)
	}
	return slices.Contains(states, state)
}

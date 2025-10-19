package store

import (
	"context"
	"fmt"
)

// get by taskid
func GetByIndex(taskId int) (TodoListItem, error) {
	if item, ok := sessionDatabase[taskId]; !ok {
		empty := TodoListItem{}
		return empty, fmt.Errorf("item not found")
	} else {
		return item, nil
	}
}

// list items
func GetList() (TodoListItems, error) {
	return sessionDatabase, nil
}

func Create(ctx context.Context, candidate TodoListItem) (TodoListItem, error) {
	taskId := AddTask(ctx, candidate.Description)
	if current, ok := sessionDatabase[taskId]; !ok {
		empty := TodoListItem{}
		return empty, fmt.Errorf("not added")
	} else {
		return current, nil
	}
}

func Update(ctx context.Context, item TodoListItem) (TodoListItem, error) {
	if _, ok := sessionDatabase[item.Line]; !ok {
		empty := TodoListItem{}
		return empty, fmt.Errorf("item not found")
	} else {
		return UpdateTask(ctx, item)
	}
}

func Delete(ctx context.Context, taskId int) error {
	if !DeleteTask(ctx, taskId) {
		return fmt.Errorf("item not found")
	} else {
		return nil
	}
}

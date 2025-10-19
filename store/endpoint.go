package store

import (
	"context"
	"fmt"

	"github.com/anthriscus/appcli/appcontext"
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

func Delete(taskId int) error {
	ctx := tempContext()
	if !DeleteTask(ctx, taskId) {
		return fmt.Errorf("item not found")
	} else {
		return nil
	}
}

// very temp situation for context obj , while we work out Handling all the context flow.
// we should already have a context at this point
func tempContext() context.Context {
	id := GenerateId()
	return context.WithValue(context.Background(), appcontext.TraceIdKey, id)
}

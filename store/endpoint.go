package store

import (
	"context"
	"fmt"
)

// get by taskid
func GetByIndex(taskId int64) (TodoListItem, error) {
	if item, ok := sessionDatabase[taskId]; !ok {
		empty := TodoListItem{}
		return empty, fmt.Errorf("item not found")
	} else {
		return item, nil
	}
}

// list items
func GetList() (TodoListItems, error) {
	items := TodoListItems{}
	keys := storeActor.Keys()
	for _, v := range keys {
		record := storeActor.Read(v)
		if record.ok {
			items[v] = record.item
		}
	}
	return items, nil
}

func Create(ctx context.Context, candidate TodoListItem) (TodoListItem, error) {
	if taskId, ok := AddTask(ctx, candidate.Description); ok != nil {
		empty := TodoListItem{}
		return empty, fmt.Errorf("not added")
	} else {
		record := storeActor.Read(taskId)
		if record.ok {
			return record.item, nil
		}
		empty := TodoListItem{}
		return empty, fmt.Errorf("not added")
	}
}

func Update(ctx context.Context, item TodoListItem) (TodoListItem, error) {
	record := storeActor.Read(item.Line)
	if record.ok {
		return UpdateTask(ctx, item)
	}
	empty := TodoListItem{}
	return empty, fmt.Errorf("item not found")
}

func Delete(ctx context.Context, taskId int64) error {
	if ok := DeleteTask(ctx, taskId); ok != nil {
		return ok
	} else {
		return ok
	}
}

package internal

import (
	"fmt"
	"math/rand"

	"github.com/anthriscus/appcli/store"
)

func dummyTaskItem() store.TodoListItem {
	return store.TodoListItem{Description: dummyTaskDescription()}
}

func dummyTaskDescription() string {
	return fmt.Sprintf("Buy %d apples for %s", rand.Intn(1e4), randomString(62))
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func GenerateDummyTasks(size int64) store.TodoListItems {
	items := store.TodoListItems{}
	for i := range size {
		items[i] = dummyTaskItem()
	}
	return items
}

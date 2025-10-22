package store

import (
	"context"
	"testing"
	"time"
)

func TestGetByIndex(t *testing.T) {
	var tests = []struct {
		description string
		newItem     TodoListItem
		addToList   bool
		want        bool
	}{
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        1,
				Id:          "1",
				Description: "Updated task description buy apples",
				State:       1,
				Created:     time.Now().UTC(),
			},
			addToList: true,
			want:      true},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        -1,
				Id:          "-1",
				Description: "Updated task description buy apples",
				State:       1,
				Created:     time.Now().UTC(),
			},
			addToList: false,
			want:      false},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        0,
				Id:          "0",
				Description: "Updated task description buy apples",
				State:       1,
				Created:     time.Now().UTC(),
			},
			addToList: false,
			want:      false},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        2,
				Id:          "2",
				Description: "Updated task description buy apples",
				State:       1,
				Created:     time.Now().UTC(),
			},
			addToList: false,
			want:      false},
	}
	resetList()
	ctx := context.Background()
	for _, tc := range tests {

		switch {
		case tc.addToList:
			if _, ok := AddTask(ctx, tc.description); ok != nil {
				t.Errorf("item %d not added to for a fetch %s\n", tc.newItem.Line, tc.description)
			} else if _, ok := GetByIndex(tc.newItem.Line); tc.want != (ok == nil) {
				t.Errorf("item %d not fetched\n", tc.newItem.Line)
			}
		case !tc.addToList:
			if _, ok := GetByIndex(tc.newItem.Line); tc.want != (ok == nil) {
				t.Errorf("item %d should be not fetched\n", tc.newItem.Line)
			}
		}
	}
}
func TestGetList(t *testing.T) {
	var tests = []struct {
		newItem   []string
		item      int
		addToList bool
		want      bool
	}{
		{
			newItem: []string{
				"Original task description buy apples", "Original task description buy apples", "Original task description buy apples", "Original task description buy apples"},
			item:      1,
			addToList: true,
			want:      true},
		{
			newItem:   []string{},
			item:      2,
			addToList: false,
			want:      true},
	}
	ctx := context.Background()
	for _, tc := range tests {
		switch {
		case tc.addToList:
			resetList()
			for _, desc := range tc.newItem {
				if _, ok := AddTask(ctx, desc); ok != nil {
					t.Errorf("items %d not added for a fetch test %s\n", tc.item, desc)
				}
			}
			if r, ok := GetList(); tc.want != (ok == nil) {
				t.Errorf("items %d not fetched\n", tc.item)
			} else if tc.want != (r != nil) {
				t.Errorf("items %d not fetched\n", tc.item)
			} else if tc.want != (len(r) > 0) {
				t.Errorf("items %d not fetched\n", tc.item)
			}
		case !tc.addToList:
			resetList()
			if r, ok := GetList(); tc.want != (ok == nil) {
				t.Errorf("items %d not fetched\n", tc.item)
			} else if tc.want != (len(r) == 0) {
				t.Errorf("items %d not fetched\n", tc.item)
			}
		}
	}
}

func TestCreate(t *testing.T) {
	var tests = []struct {
		newItem TodoListItem
		want    bool
	}{
		{newItem: TodoListItem{
			Line:        1,
			Id:          "1",
			Description: "Original task description buy apples",
			State:       1,
			Created:     time.Now().UTC(),
		},
			want: true},
		{newItem: TodoListItem{
			Line:        2,
			Id:          "2",
			Description: "", // bad description
			State:       1,
			Created:     time.Now().UTC(),
		},
			want: false},
	}
	resetList()
	ctx := context.Background()
	for _, tc := range tests {
		if _, ok := Create(ctx, tc.newItem); tc.want != (ok == nil) {
			t.Errorf("item %d not added %s\n", tc.newItem.Line, tc.newItem.Description)
		}
	}
}

func TestUpdate(t *testing.T) {
	var tests = []struct {
		description string
		newItem     TodoListItem
		want        bool
	}{
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        1,
				Id:          "1",
				Description: "Updated task description buy apples",
				State:       1,
				Created:     time.Now().UTC(),
			},
			want: true},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        2,
				Id:          "2",
				Description: "", // bad description
				State:       1,
				Created:     time.Now().UTC(),
			},
			want: false},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        3,
				Id:          "3",
				Description: "Updated task description buy apples",
				State:       -1, // bad state
				Created:     time.Now().UTC(),
			},
			want: false},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        4,
				Id:          "4",
				Description: "Updated task description buy apples",
				State:       1,
				Created:     time.Now().UTC(),
			},
			want: true},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        5,
				Id:          "5",
				Description: "Updated task description buy apples",
				State:       2,
				Created:     time.Now().UTC(),
			},
			want: true},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        6,
				Id:          "6",
				Description: "Updated task description buy apples",
				State:       3, // bad state
				Created:     time.Now().UTC(),
			},
			want: false},
	}
	resetList()
	ctx := context.Background()
	for _, tc := range tests {
		if _, ok := AddTask(ctx, tc.description); ok != nil {
			t.Errorf("item %d not added to for a test change %s\n", tc.newItem.Line, tc.description)
		} else if _, ok := Update(ctx, tc.newItem); tc.want != (ok == nil) {
			t.Errorf("item %d not changd %s\n", tc.newItem.Line, tc.newItem.Description)
		}
	}
}

func TestDelete(t *testing.T) {
	var tests = []struct {
		description string
		item        int
		addToList   bool
		want        bool
	}{
		{description: "Original task description buy apples",
			item:      1,
			addToList: true,
			want:      true},
		{description: "Original task description buy apples",
			item:      -1,
			addToList: false,
			want:      false},
		{description: "Original task description buy apples",
			item:      0,
			addToList: false,
			want:      false},
		{description: "Original task description buy apples",
			item:      2,
			addToList: false,
			want:      false},
	}

	ctx := context.Background()
	for _, tc := range tests {
		switch {
		case tc.addToList:
			resetList()
			if _, ok := AddTask(ctx, tc.description); ok != nil {
				t.Errorf("items %d not added for a delete test %s\n", tc.item, tc.description)
			} else if ok := DeleteTask(ctx, tc.item); tc.want != ok {
				t.Errorf("item %d not deleted %s\n", tc.item, tc.description)
			}
		case !tc.addToList:
			resetList()
			if ok := DeleteTask(ctx, tc.item); tc.want != ok {
				t.Errorf("item %d not deleted %s\n", tc.item, tc.description)
			}
		}
	}
}

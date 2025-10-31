package store

import (
	"testing"
	"time"

	"github.com/anthriscus/appcli/logging"
)

func TestMain(m *testing.M) {
	logging.Default()
	logging.Log().Info("setup default logging to std.io for tests")
	resetList()

	m.Run()
}

func TestIsState(t *testing.T) {
	state := StateNotStarted
	want := true

	got := isState((state))
	if got != want {
		t.Errorf(`isState() = %t, want %t`, got, want)
	}
}

func TestIsStateAll(t *testing.T) {
	var tests = []struct {
		state int
		want  bool
	}{
		{
			state: StateNotStarted,
			want:  true},
		{
			state: StateStarted,
			want:  true},
		{
			state: StateCompleted,
			want:  true},
		{
			state: -1,
			want:  false},
		{
			state: 1,
			want:  true},
		{
			state: 2,
			want:  true},
		{
			state: 3,
			want:  false},
		{
			state: 4,
			want:  false},
	}
	//
	for i, tc := range tests {
		got := isState((tc.state))
		if got != tc.want {
			t.Errorf(`check item:%d isState() = %t, want %t`, i, got, tc.want)
		}
	}
}

func TestAddTask(t *testing.T) {
	var tests = []struct {
		description string
		want        bool
	}{
		{description: "Original task description buy apples",
			want: true},
		{description: "",
			want: false},
	}
	ctx := t.Context()
	StartActor(ctx)
	resetList()

	for _, tc := range tests {
		if r, ok := AddTask(ctx, tc.description); tc.want != (ok == nil) {
			t.Errorf("not added %s\n", tc.description)
		} else if v, ok := currentList()[r]; tc.want != ok {
			t.Errorf("not added %s\n", tc.description)
		} else if v.Description != tc.description {
			t.Errorf("not added %s\n", tc.description)
		}
	}
}
func TestDescriptionChange(t *testing.T) {
	var tests = []struct {
		description    string
		newDescription string
		index          int
		want           bool
	}{
		{description: "Original task description buy apples",
			newDescription: "Updated task description buy apples",
			index:          1,
			want:           true},
		{description: "Original task description buy pears",
			newDescription: "",
			index:          2,
			want:           false},
	}
	ctx := t.Context()
	StartActor(ctx)
	resetList()

	for _, tc := range tests {
		if added, ok := AddTask(ctx, tc.description); ok != nil {
			t.Errorf("item %d not added to for a change %s\n", tc.index, tc.description)
		} else if ok := DescriptionChange(ctx, added, tc.newDescription); tc.want != (ok == nil) {
			t.Errorf("item %d not changd %s\n", tc.index, tc.newDescription)
		}
	}
}

func TestUpdateTask(t *testing.T) {
	var tests = []struct {
		description string
		newItem     TodoListItem
		want        bool
	}{
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        1,
				Description: "Updated task description buy apples",
				State:       1,
				Created:     time.Now().UTC(),
			},
			want: true},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        2,
				Description: "", // bad description
				State:       1,
				Created:     time.Now().UTC(),
			},
			want: false},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        3,
				Description: "Updated task description buy apples",
				State:       -1, // bad state
				Created:     time.Now().UTC(),
			},
			want: false},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        4,
				Description: "Updated task description buy apples",
				State:       1,
				Created:     time.Now().UTC(),
			},
			want: true},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        5,
				Description: "Updated task description buy apples",
				State:       2,
				Created:     time.Now().UTC(),
			},
			want: true},
		{description: "Original task description buy apples",
			newItem: TodoListItem{
				Line:        6,
				Description: "Updated task description buy apples",
				State:       3, // bad state
				Created:     time.Now().UTC(),
			},
			want: false},
	}
	ctx := t.Context()
	StartActor(ctx)
	resetList()

	for _, tc := range tests {
		if added, ok := AddTask(ctx, tc.description); ok != nil {
			t.Errorf("item %d not added to for a test change %s\n", tc.newItem.Line, tc.description)
		} else if _, ok := UpdateTask(ctx, TodoListItem{Line: added, Description: tc.newItem.Description, State: tc.newItem.State}); tc.want != (ok == nil) {
			t.Errorf("item %d not changd %s\n", tc.newItem.Line, tc.newItem.Description)
		}
	}
}

func TestDeleteTask(t *testing.T) {
	var tests = []struct {
		description string
		item        int64
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
	ctx := t.Context()
	StartActor(ctx)

	for _, tc := range tests {
		switch {
		case tc.addToList:
			resetList()
			if added, ok := AddTask(ctx, tc.description); ok != nil {
				t.Errorf("items %d not added for a delete test %s\n", tc.item, tc.description)
			} else if ok := DeleteTask(ctx, added); tc.want != (ok == nil) {
				t.Errorf("item %d not deleted %s\n", tc.item, tc.description)
			}
		case !tc.addToList:
			resetList()
			if ok := DeleteTask(ctx, tc.item); tc.want != (ok == nil) {
				t.Errorf("item %d not deleted %s\n", tc.item, tc.description)
			}
		}
	}
}

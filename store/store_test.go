package store

import (
	"bytes"
	"context"
	"testing"

	"github.com/anthriscus/appcli/appcontext"
	"github.com/anthriscus/appcli/logging"
)

func TestMain(m *testing.M) {
	logging.Default()
	logging.Log().Info("setup default logging to std.io for tests")
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

func TestOpenSession(t *testing.T) {
	dir := t.TempDir()
	var tests = []struct {
		datafile string
		want     bool
	}{
		{
			datafile: dir + "\\testData.json",
			want:     true},
		{
			datafile: dir + "\\",
			want:     false},
	}
	ctx := context.Background()

	for i, tc := range tests {
		if ok := OpenSession(ctx, tc.datafile); tc.want != (ok == nil) {
			t.Errorf("test %d want: %t got: %t", i, tc.want, (ok == nil))
		} else {
			logging.Log().Info("test looking:", "i", i, "want", tc.want, "got", (ok == nil))
		}
	}
}

func TestRestoreList(t *testing.T) {
	ctx := context.WithValue(context.Background(), appcontext.TraceIdKey, "42")
	var tests = []struct {
		json string
		want bool
	}{
		{json: "{\"42\": {\"line\": 42,\"description\": \"Build awesome new app with Go\",\"state\": 1,\"created\": \"2025-10-10T01:00:00.0000000Z\",\"id\": \"1234\"} }",
			want: false},
		{
			json: "{}",
			want: false},
		{json: "",
			want: false},
		{json: "acme task",
			want: true},
		{json: "\"acme task\"",
			want: true},
		{json: "\"1\": {\"line\": 42,\"description\":",
			want: true},
	}

	// assert as a todolist
	for i, tc := range tests {
		b := new(bytes.Buffer)
		b.WriteString(tc.json)
		// check a restore
		logging.Log().Info("CHECKING:", "json", tc.json)
		if v, ok := restoreList(ctx, b); tc.want != (ok != nil) {
			t.Errorf("restoring list from json , %d got error on item  %+v", i, tc.json)
		} else if v == nil {
			t.Errorf("restoring list from json , %d want empty item got nil %+v", i, tc.json)
		}
	}
}

package store

import (
	"bytes"
	"context"
	"testing"

	"github.com/anthriscus/appcli/appcontext"
	"github.com/anthriscus/appcli/logging"
)

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

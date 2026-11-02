package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestCmdFolderInfo(t *testing.T) {
	got := testWithServer(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/2/paper/docs/get_folder_info" {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			var v map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&v); err != nil || v["doc_id"] != "xxx" {
				t.Errorf("bad request: %v %v", v, err)
			}
			fmt.Fprintln(w, `{"folders":[{"id":"f1","name":"foo"},{"id":"f2","name":"bar"}]}`)
		},
		func(app *cli.App) {
			if err := app.Run([]string{"dboxpaper", "folder_info", "xxx"}); err != nil {
				t.Error(err)
			}
		},
	)
	want := "f1 foo\nf2 bar\n"
	if got != want {
		t.Fatalf("want %q but got %q", want, got)
	}
}

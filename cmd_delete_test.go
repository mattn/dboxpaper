package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestCmdDelete(t *testing.T) {
	got := testWithServer(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/2/paper/docs/permanently_delete":
				var v map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&v)
				if err != nil {
					t.Error(err)
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				if v["doc_id"] != "xxx" {
					t.Error("bad request")
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		},
		func(app *cli.App) {
			if err := app.Run([]string{"dboxpaper", "delete", "xxx"}); err != nil {
				t.Error(err)
			}
		},
	)
	if got != "" {
		t.Fatalf("want %v but got %v", "", got)
	}
}

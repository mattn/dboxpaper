package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/urfave/cli"
)

func TestCmdDelete(t *testing.T) {
	got := testWithServer(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/2/paper/docs/permanently_delete":
				var v map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&v)
				if err != nil {
					t.Fatal(err)
				}
				if v["doc_id"] != "xxx" {
					t.Fatal("bad request")
				}
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		},
		func(app *cli.App) {
			app.Run([]string{"dboxpaper", "delete", "xxx"})
		},
	)
	if got != "" {
		t.Fatalf("want %v but got %v", "", got)
	}
}

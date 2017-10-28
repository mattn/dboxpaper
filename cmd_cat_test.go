package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/urfave/cli"
)

func TestCmdCat(t *testing.T) {
	got := testWithServer(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/2/paper/docs/download":
				h := r.Header.Get("Dropbox-API-Arg")
				var v map[string]interface{}
				err := json.NewDecoder(strings.NewReader(h)).Decode(&v)
				if err != nil {
					t.Fatal(err)
				}
				if v["doc_id"] != "xxx" {
					t.Fatal("bad request")
				}
				fmt.Fprintln(w, `hello world`)
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		},
		func(app *cli.App) {
			app.Run([]string{"dboxpaper", "cat", "xxx"})
		},
	)
	want := "hello world\n"
	if got != want {
		t.Fatalf("want %v but got %v", want, got)
	}
}

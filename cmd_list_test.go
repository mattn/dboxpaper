package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestCmdList(t *testing.T) {
	got := testWithServer(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/2/paper/docs/list":
				fmt.Fprintln(w, `{"doc_ids":["xxx","yyy"]}`)
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		},
		func(app *cli.App) {
			app.Run([]string{"dboxpaper", "ls"})
		},
	)
	want := "xxx\nyyy\n"
	if got != want {
		t.Fatalf("want %v but got %v", want, got)
	}
}

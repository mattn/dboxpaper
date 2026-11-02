package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestCmdList(t *testing.T) {
	got := testWithServer(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/2/users/features/get_values":
				fmt.Fprintln(w, `{"values":[{".tag":"paper_as_files","paper_as_files":{".tag":"enabled","enabled":false}}]}`)
				return
			case "/2/paper/docs/list":
				fmt.Fprintln(w, `{"doc_ids":["xxx","yyy"]}`)
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		},
		func(app *cli.App) {
			if err := app.Run([]string{"dboxpaper", "ls"}); err != nil {
				t.Error(err)
			}
		},
	)
	want := "xxx\nyyy\n"
	if got != want {
		t.Fatalf("want %v but got %v", want, got)
	}
}

func listTestHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/2/users/features/get_values":
			fmt.Fprintln(w, `{"values":[{".tag":"paper_as_files","paper_as_files":{".tag":"enabled","enabled":false}}]}`)
			return
		case "/2/paper/docs/list":
			fmt.Fprintln(w, `{"doc_ids":["xxx"],"cursor":{"value":"c1"},"has_more":true}`)
			return
		case "/2/paper/docs/list/continue":
			var v map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&v); err != nil || v["cursor"] != "c1" {
				t.Errorf("bad cursor: %v %v", v, err)
			}
			fmt.Fprintln(w, `{"doc_ids":["yyy"],"has_more":false}`)
			return
		case "/2/paper/docs/get_metadata":
			var v map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
				t.Error(err)
			}
			if v["doc_id"] == "yyy" {
				http.Error(w, `{"error_summary":"doc_not_found/"}`, http.StatusConflict)
				return
			}
			fmt.Fprintf(w, `{"doc_id":%q,"title":"title of %s"}`+"\n", v["doc_id"], v["doc_id"])
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func TestCmdListPagination(t *testing.T) {
	got := testWithServer(listTestHandler(t), func(app *cli.App) {
		if err := app.Run([]string{"dboxpaper", "ls"}); err != nil {
			t.Error(err)
		}
	})
	want := "xxx\nyyy\n"
	if got != want {
		t.Fatalf("want %q but got %q", want, got)
	}
}

func TestCmdListTitle(t *testing.T) {
	var errbuf bytes.Buffer
	got := testWithServer(listTestHandler(t), func(app *cli.App) {
		app.ErrWriter = &errbuf
		defer func() { app.ErrWriter = nil }()
		if err := app.Run([]string{"dboxpaper", "ls", "--title"}); err != nil {
			t.Error(err)
		}
	})
	want := "xxx title of xxx\n"
	if got != want {
		t.Fatalf("want %q but got %q", want, got)
	}
	if !strings.Contains(errbuf.String(), "yyy") {
		t.Fatalf("want warning for yyy but got %q", errbuf.String())
	}
}

func TestCmdListJSON(t *testing.T) {
	got := testWithServer(listTestHandler(t), func(app *cli.App) {
		app.ErrWriter = io.Discard
		defer func() { app.ErrWriter = nil }()
		if err := app.Run([]string{"dboxpaper", "ls", "--json"}); err != nil {
			t.Error(err)
		}
	})
	want := `[{"doc_id":"xxx","title":"title of xxx"}]` + "\n"
	if got != want {
		t.Fatalf("want %q but got %q", want, got)
	}
}

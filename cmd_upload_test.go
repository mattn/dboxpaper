package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestCmdUploadCreate(t *testing.T) {
	got := testWithServer(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/2/paper/docs/create":
				h := r.Header.Get("Dropbox-API-Arg")
				var v map[string]interface{}
				err := json.NewDecoder(strings.NewReader(h)).Decode(&v)
				if err != nil {
					t.Error(err)
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				if _, ok := v["doc_id"]; ok {
					t.Error("bad request")
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				if string(b) != "hello world" {
					t.Error("bad request")
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				fmt.Fprintln(w, `{"doc_id":"xxx"}`)
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		},
		func(app *cli.App) {
			app.Metadata["stdin"] = strings.NewReader("hello world")
			if err := app.Run([]string{"dboxpaper", "upload"}); err != nil {
				t.Error(err)
			}
		},
	)
	want := "xxx\n"
	if got != want {
		t.Fatalf("want %v but got %v", want, got)
	}
}

func TestCmdUploadUpdate(t *testing.T) {
	got := testWithServer(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/2/paper/docs/get_metadata":
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
				fmt.Fprintln(w, `{"doc_id":"xxx","revision":123}`)
				return
			case "/2/paper/docs/update":
				h := r.Header.Get("Dropbox-API-Arg")
				var v map[string]interface{}
				err := json.NewDecoder(strings.NewReader(h)).Decode(&v)
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
				if v["revision"] != float64(123) {
					t.Error("bad request")
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				if string(b) != "good morning" {
					t.Error("bad request")
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				fmt.Fprintln(w, `{"doc_id":"xxx"}`)
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		},
		func(app *cli.App) {
			app.Metadata["stdin"] = strings.NewReader("good morning")
			if err := app.Run([]string{"dboxpaper", "upload", "xxx"}); err != nil {
				t.Error(err)
			}
		},
	)
	want := "xxx\n"
	if got != want {
		t.Fatalf("want %v but got %v", want, got)
	}
}

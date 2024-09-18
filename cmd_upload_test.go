package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
					t.Fatal(err)
				}
				if _, ok := v["doc_id"]; ok {
					t.Fatal("bad request")
				}
				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				if string(b) != "hello world" {
					t.Fatal("bad request")
				}
				fmt.Fprintln(w, `{"doc_id":"xxx"}`)
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		},
		func(app *cli.App) {
			app.Metadata["stdin"] = strings.NewReader("hello world")
			app.Run([]string{"dboxpaper", "upload"})
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
					t.Fatal(err)
				}
				if v["doc_id"].(string) != "xxx" {
					t.Fatal("bad request")
				}
				fmt.Fprintln(w, `{"doc_id":"xxx","revision":123}`)
				return
			case "/2/paper/docs/update":
				h := r.Header.Get("Dropbox-API-Arg")
				var v map[string]interface{}
				err := json.NewDecoder(strings.NewReader(h)).Decode(&v)
				if err != nil {
					t.Fatal(err)
				}
				if v["doc_id"].(string) != "xxx" {
					t.Fatal("bad request")
				}
				if int64(v["revision"].(float64)) != 123 {
					t.Fatal("bad request")
				}
				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				if string(b) != "good morning" {
					t.Fatal("bad request")
				}
				fmt.Fprintln(w, `{"doc_id":"xxx"}`)
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		},
		func(app *cli.App) {
			app.Metadata["stdin"] = strings.NewReader("good morning")
			app.Run([]string{"dboxpaper", "upload", "xxx"})
		},
	)
	want := "xxx\n"
	if got != want {
		t.Fatalf("want %v but got %v", want, got)
	}
}

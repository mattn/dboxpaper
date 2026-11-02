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

// fakeFiles is a fake Dropbox files API for a "Paper as files" account.
type fakeFiles struct {
	t     *testing.T
	files map[string]string // path -> markdown
	calls []string
}

func (f *fakeFiles) arg(r *http.Request) map[string]interface{} {
	var v map[string]interface{}
	src := r.Header.Get("Dropbox-API-Arg")
	if src == "" {
		b, _ := io.ReadAll(r.Body)
		src = string(b)
	}
	if err := json.Unmarshal([]byte(src), &v); err != nil {
		f.t.Errorf("%s: bad argument %q: %v", r.URL.Path, src, err)
	}
	return v
}

func notFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	fmt.Fprintln(w, `{"error_summary":"path/not_found/"}`)
}

func (f *fakeFiles) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.calls = append(f.calls, r.URL.Path)
	switch r.URL.Path {
	case "/2/users/features/get_values":
		fmt.Fprintln(w, `{"values":[{".tag":"paper_as_files","paper_as_files":{".tag":"enabled","enabled":true}}]}`)
	case "/2/files/list_folder":
		v := f.arg(r)
		if v["path"] != "" || v["recursive"] != true {
			f.t.Errorf("bad list_folder argument: %v", v)
		}
		fmt.Fprintln(w, `{"entries":[
			{".tag":"folder","name":"Migrated Paper Docs","path_display":"/Migrated Paper Docs"},
			{".tag":"file","name":"メモ.paper","id":"id:a","path_display":"/Migrated Paper Docs/メモ.paper"},
			{".tag":"file","name":"photo.jpg","id":"id:b","path_display":"/photo.jpg"}
		],"cursor":"c1","has_more":true}`)
	case "/2/files/list_folder/continue":
		if v := f.arg(r); v["cursor"] != "c1" {
			f.t.Errorf("bad cursor: %v", v)
		}
		fmt.Fprintln(w, `{"entries":[
			{".tag":"file","name":"todo.paper","id":"id:c","path_display":"/todo.paper"}
		],"cursor":"c2","has_more":false}`)
	case "/2/files/export":
		v := f.arg(r)
		if v["export_format"] != "markdown" {
			f.t.Errorf("bad export_format: %v", v)
		}
		body, ok := f.files[v["path"].(string)]
		if !ok {
			notFound(w)
			return
		}
		fmt.Fprint(w, body)
	case "/2/files/get_metadata":
		path := f.arg(r)["path"].(string)
		if _, ok := f.files[path]; !ok {
			notFound(w)
			return
		}
		name := path[strings.LastIndex(path, "/")+1:]
		fmt.Fprintf(w, `{".tag":"file","name":%q,"path_display":%q}`+"\n", name, path)
	case "/2/files/paper/create":
		v := f.arg(r)
		if v["import_format"] != "markdown" {
			f.t.Errorf("bad import_format: %v", v)
		}
		b, _ := io.ReadAll(r.Body)
		path := v["path"].(string)
		f.files[path] = string(b)
		fmt.Fprintf(w, `{"result_path":%q,"file_id":"id:new","paper_revision":1}`+"\n", path)
	case "/2/files/paper/update":
		v := f.arg(r)
		if v["doc_update_policy"] != "overwrite" {
			f.t.Errorf("bad doc_update_policy: %v", v)
		}
		b, _ := io.ReadAll(r.Body)
		f.files[v["path"].(string)] = string(b)
		fmt.Fprintln(w, `{"paper_revision":2}`)
	case "/2/files/delete_v2":
		path := f.arg(r)["path"].(string)
		if _, ok := f.files[path]; !ok {
			notFound(w)
			return
		}
		delete(f.files, path)
		fmt.Fprintln(w, `{"metadata":{}}`)
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func runFiles(t *testing.T, f *fakeFiles, stdin string, args ...string) (string, error) {
	var err error
	got := testWithServer(f.ServeHTTP, func(app *cli.App) {
		app.Metadata["stdin"] = strings.NewReader(stdin)
		err = app.Run(append([]string{"dboxpaper"}, args...))
	})
	return got, err
}

func newFakeFiles(t *testing.T) *fakeFiles {
	return &fakeFiles{t: t, files: map[string]string{"/todo.paper": "# todo\n"}}
}

func TestPaperFilesList(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"ls"}, "/Migrated Paper Docs/メモ.paper\n/todo.paper\n"},
		{[]string{"ls", "--title"}, "/Migrated Paper Docs/メモ.paper メモ\n/todo.paper todo\n"},
	}
	for _, tt := range tests {
		got, err := runFiles(t, newFakeFiles(t), "", tt.args...)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("%v: want %q but got %q", tt.args, tt.want, got)
		}
	}

	got, err := runFiles(t, newFakeFiles(t), "", "ls", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var v []FileMeta
	if err := json.Unmarshal([]byte(got), &v); err != nil {
		t.Fatal(err)
	}
	if len(v) != 2 || v[1].ID != "id:c" {
		t.Fatalf("unexpected JSON output: %s", got)
	}
}

func TestPaperFilesCat(t *testing.T) {
	got, err := runFiles(t, newFakeFiles(t), "", "cat", "/todo.paper")
	if err != nil {
		t.Fatal(err)
	}
	if got != "# todo\n" {
		t.Fatalf("want %q but got %q", "# todo\n", got)
	}

	_, err = runFiles(t, newFakeFiles(t), "", "cat", "/missing.paper")
	if !isNotFound(err) {
		t.Fatalf("want not_found error but got %v", err)
	}
}

func TestPaperFilesUpload(t *testing.T) {
	f := newFakeFiles(t)
	got, err := runFiles(t, f, "# new\n", "upload", "/日本語.paper")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/日本語.paper\n" || f.files["/日本語.paper"] != "# new\n" {
		t.Fatalf("create failed: %q %v", got, f.files)
	}

	got, err = runFiles(t, f, "# updated\n", "upload", "/todo.paper")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/todo.paper\n" || f.files["/todo.paper"] != "# updated\n" {
		t.Fatalf("update failed: %q %v", got, f.files)
	}

	_, err = runFiles(t, f, "x", "upload", "/new.txt")
	if err == nil || !strings.Contains(err.Error(), ".paper") {
		t.Fatalf("want error for non .paper path but got %v", err)
	}
}

func TestPaperFilesDelete(t *testing.T) {
	f := newFakeFiles(t)
	if _, err := runFiles(t, f, "", "delete", "/todo.paper"); err != nil {
		t.Fatal(err)
	}
	if _, ok := f.files["/todo.paper"]; ok {
		t.Fatal("file was not deleted")
	}
}

func TestPaperFilesFolderInfo(t *testing.T) {
	_, err := runFiles(t, newFakeFiles(t), "", "folder_info", "/todo.paper")
	if err == nil {
		t.Fatal("want error but got nil")
	}
}

func TestListFilesFlag(t *testing.T) {
	// --files must not need the features endpoint.
	f := newFakeFiles(t)
	if _, err := runFiles(t, f, "", "ls", "--files"); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.calls {
		if c == "/2/users/features/get_values" {
			t.Fatal("features endpoint should not be called with --files")
		}
	}
}

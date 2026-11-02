package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestCmdErrorPropagation(t *testing.T) {
	tests := [][]string{
		{"dboxpaper", "cat", "xxx"},
		{"dboxpaper", "delete", "xxx"},
		{"dboxpaper", "folder_info", "xxx"},
		{"dboxpaper", "ls"},
	}
	for _, args := range tests {
		var err error
		testWithServer(
			func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, `{"error_summary":"doc_not_found/"}`, http.StatusConflict)
			},
			func(app *cli.App) {
				err = app.Run(args)
			},
		)
		if err == nil || !strings.Contains(err.Error(), "doc_not_found") {
			t.Errorf("%v: want doc_not_found error but got %v", args[1:], err)
		}
	}
}

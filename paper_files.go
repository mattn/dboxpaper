package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Accounts migrated to "Paper as files" keep their docs as .paper files in
// Dropbox, which are handled with the files API instead of the Paper API.

// FileMeta is struct for dropbox files API
type FileMeta struct {
	Tag            string `json:".tag"`
	Name           string `json:"name"`
	ID             string `json:"id"`
	PathDisplay    string `json:"path_display"`
	ServerModified string `json:"server_modified"`
}

// Title returns the file name without .paper extension.
func (f *FileMeta) Title() string {
	if !isPaperFile(f.Name) {
		return f.Name
	}
	return f.Name[:len(f.Name)-len(".paper")]
}

func isPaperFile(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".paper")
}

// isFilePath reports whether arg names a .paper file (path or file ID)
// rather than a legacy Paper doc ID.
func isFilePath(arg string) bool {
	return strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, "id:")
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not_found")
}

// paperAsFiles reports whether the account stores Paper docs as files.
func (dboxpaper *DboxPaper) paperAsFiles(ctx context.Context) (bool, error) {
	var res struct {
		Values []struct {
			Tag          string `json:".tag"`
			PaperAsFiles struct {
				Enabled bool `json:"enabled"`
			} `json:"paper_as_files"`
		} `json:"values"`
	}
	err := dboxpaper.rpc(ctx, "/2/users/features/get_values",
		map[string]interface{}{"features": []map[string]string{{".tag": "paper_as_files"}}}, &res)
	if err != nil {
		return false, err
	}
	for _, v := range res.Values {
		if v.Tag == "paper_as_files" {
			return v.PaperAsFiles.Enabled, nil
		}
	}
	return false, nil
}

// listPaperFiles returns all .paper files in the account. It uses the search
// API since walking the whole Dropbox with list_folder is too slow.
func (dboxpaper *DboxPaper) listPaperFiles(ctx context.Context) ([]FileMeta, []json.RawMessage, error) {
	var files []FileMeta
	var raws []json.RawMessage
	path := "/2/files/search_v2"
	var arg interface{} = map[string]interface{}{
		"query": ".paper",
		"options": map[string]interface{}{
			"file_categories": []string{"paper"},
			"max_results":     1000,
		},
	}
	for {
		var res struct {
			Matches []struct {
				Metadata struct {
					Metadata json.RawMessage `json:"metadata"`
				} `json:"metadata"`
			} `json:"matches"`
			Cursor  string `json:"cursor"`
			HasMore bool   `json:"has_more"`
		}
		if err := dboxpaper.rpc(ctx, path, arg, &res); err != nil {
			return nil, nil, err
		}
		for _, m := range res.Matches {
			raw := m.Metadata.Metadata
			var f FileMeta
			if err := json.Unmarshal(raw, &f); err != nil {
				return nil, nil, err
			}
			if f.Tag == "file" && isPaperFile(f.Name) {
				files = append(files, f)
				raws = append(raws, raw)
			}
		}
		if !res.HasMore {
			return files, raws, nil
		}
		path = "/2/files/search/continue_v2"
		arg = map[string]interface{}{"cursor": res.Cursor}
	}
}

func (dboxpaper *DboxPaper) exportPaperFile(ctx context.Context, path string, w io.Writer) error {
	return dboxpaper.doAPI(ctx, http.MethodPost, "/2/files/export", &request{
		content: true,
		arg:     map[string]interface{}{"path": path, "export_format": "markdown"},
		out:     w,
	})
}

// uploadPaperFile overwrites the .paper file at path, or creates it if it
// does not exist. It returns the path of the file.
func (dboxpaper *DboxPaper) uploadPaperFile(ctx context.Context, path string, r io.Reader) (string, error) {
	var meta FileMeta
	err := dboxpaper.rpc(ctx, "/2/files/get_metadata", map[string]interface{}{"path": path}, &meta)
	if err == nil {
		if meta.Tag != "file" || !isPaperFile(meta.Name) {
			return "", fmt.Errorf("%s is not a .paper file", path)
		}
		err = dboxpaper.doAPI(ctx, http.MethodPost, "/2/files/paper/update", &request{
			content: true,
			ct:      "application/octet-stream",
			arg: map[string]interface{}{
				"path":              path,
				"import_format":     "markdown",
				"doc_update_policy": "overwrite",
			},
			in: r,
		})
		if err != nil {
			return "", err
		}
		return meta.PathDisplay, nil
	}
	if !isNotFound(err) {
		return "", err
	}
	if !strings.HasPrefix(path, "/") || !isPaperFile(path) {
		return "", errors.New("path of new paper must start with / and end with .paper")
	}

	var out bytes.Buffer
	err = dboxpaper.doAPI(ctx, http.MethodPost, "/2/files/paper/create", &request{
		content: true,
		ct:      "application/octet-stream",
		arg:     map[string]interface{}{"path": path, "import_format": "markdown"},
		in:      r,
		out:     &out,
	})
	if err != nil {
		return "", err
	}
	var res struct {
		ResultPath string `json:"result_path"`
	}
	if err := json.NewDecoder(&out).Decode(&res); err != nil {
		return "", err
	}
	return res.ResultPath, nil
}

func (dboxpaper *DboxPaper) deletePaperFile(ctx context.Context, path string) error {
	return dboxpaper.rpc(ctx, "/2/files/delete_v2", map[string]interface{}{"path": path}, nil)
}

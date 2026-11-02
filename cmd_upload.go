package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/urfave/cli/v2"
)

func init() {
	command := &cli.Command{
		Name:      "upload",
		Aliases:   []string{"up"},
		Usage:     "Upload paper",
		ArgsUsage: "[doc_id or /path/to/file.paper]",
		Action: func(c *cli.Context) error {
			stdin := app.Metadata["stdin"].(io.Reader)
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			if isFilePath(c.Args().First()) {
				path, err := dboxpaper.uploadPaperFile(context.Background(), c.Args().First(), stdin)
				if err != nil {
					return err
				}
				fmt.Fprintln(c.App.Writer, path)
				return nil
			}
			path := "/2/paper/docs/create"
			arg := map[string]interface{}{"import_format": "markdown"}
			if c.Args().Present() {
				path = "/2/paper/docs/update"
				arg["doc_id"] = c.Args().First()

				docmeta, _, err := dboxpaper.getMetadata(context.Background(), c.Args().First())
				if err != nil {
					return err
				}
				arg["doc_update_policy"] = "overwrite_all"
				arg["revision"] = docmeta.Revision
			}
			var out bytes.Buffer
			err := dboxpaper.doAPI(
				context.Background(),
				http.MethodPost,
				path,
				&request{
					ct:  "application/octet-stream",
					arg: arg,
					in:  stdin,
					out: &out,
				})
			if err != nil {
				return err
			}
			var docmeta DocsMeta
			err = json.NewDecoder(&out).Decode(&docmeta)
			if err != nil {
				return err
			}
			fmt.Fprintln(c.App.Writer, docmeta.DocID)
			return nil
		},
	}
	command.BashComplete = docIDCompletion
	app.Commands = append(app.Commands, command)
}

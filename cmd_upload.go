package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/urfave/cli"
)

func init() {
	command := cli.Command{
		Name:    "upload",
		Aliases: []string{"up"},
		Usage:   "Upload paper",
		Action: func(c *cli.Context) error {
			stdin := app.Metadata["stdin"].(io.Reader)
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(map[string]string{"doc_id": c.Args().First(), "export_format": "markdown"})
			if err != nil {
				return err
			}
			path := "/2/paper/docs/create"
			arg := map[string]interface{}{"import_format": "markdown"}
			if c.Args().Present() {
				path = "/2/paper/docs/update"
				arg["doc_id"] = c.Args().First()

				var in, out bytes.Buffer
				err = json.NewEncoder(&in).Encode(map[string]interface{}{"doc_id": arg["doc_id"]})
				if err != nil {
					return err
				}
				err = dboxpaper.doAPI(
					context.Background(),
					http.MethodPost,
					"/2/paper/docs/get_metadata",
					&request{
						ct:  "application/json",
						in:  &in,
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
				arg["doc_update_policy"] = "overwrite_all"
				arg["revision"] = docmeta.Revision
			}
			var out bytes.Buffer
			err = dboxpaper.doAPI(
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
	command.BashComplete = docIdCompletion
	app.Commands = append(app.Commands, command)
}

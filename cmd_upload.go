package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/urfave/cli"
)

func init() {
	command := cli.Command{
		Name:    "upload",
		Aliases: []string{"up"},
		Usage:   "Upload paper",
		Action: func(c *cli.Context) error {
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
			}
			var meta map[string]interface{}
			err = dboxpaper.doAPI(
				context.Background(),
				http.MethodPost,
				path,
				&request{
					ct:   "application/octet-stream",
					arg:  arg,
					in:   os.Stdin,
					out:  os.Stdout,
					meta: meta,
				})
			if err != nil {
				return err
			}
			fmt.Fprintln(c.App.Writer, meta["doc_id"])
			return nil
		},
	}
	command.BashComplete = docIdCompletion
	app.Commands = append(app.Commands, command)
}

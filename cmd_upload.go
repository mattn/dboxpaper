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
			var meta map[string]interface{}
			err = dboxpaper.doAPI(
				context.Background(),
				http.MethodPost,
				"https://api.dropboxapi.com/2/paper/docs/create",
				&request{
					ct:   "application/octet-stream",
					arg:  map[string]interface{}{"import_format": "markdown"},
					in:   os.Stdin,
					out:  os.Stdout,
					meta: meta,
				})
			if err != nil {
				return err
			}
			fmt.Println(meta["doc_id"])
			return nil
		},
	}
	app.Commands = append(app.Commands, command)
}

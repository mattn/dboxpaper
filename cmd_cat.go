package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli"
)

func init() {
	command := cli.Command{
		Name:    "cat",
		Aliases: []string{},
		Usage:   "Cat paper",
		Action: func(c *cli.Context) error {
			if !c.Args().Present() {
				cli.ShowCommandHelp(c, "cat")
				return nil
			}
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(map[string]string{"doc_id": c.Args().First(), "export_format": "markdown"})
			if err != nil {
				return err
			}
			return dboxpaper.doAPI(context.Background(), http.MethodPost, "https://api.dropboxapi.com/2/paper/docs/download", strings.TrimSpace(buf.String()), os.Stdout, nil)
		},
	}
	app.Commands = append(app.Commands, command)
}

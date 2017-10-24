package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/urfave/cli"
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:    "cat",
		Aliases: []string{"c"},
		Usage:   "Cat paper",
		Action: func(c *cli.Context) error {
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(map[string]string{"doc_id": c.Args().First(), "export_format": "markdown"})
			if err != nil {
				return err
			}
			var body string
			err = dboxpaper.doAPI(context.Background(), http.MethodPost, "https://api.dropboxapi.com/2/paper/docs/download", strings.TrimSpace(buf.String()), &body)
			if err != nil {
				return err
			}
			fmt.Print(body)
			println("=")
			return nil
		},
	})
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/urfave/cli"
)

func init() {
	command := cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "Show papers",
		Action: func(c *cli.Context) error {
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			var buf bytes.Buffer
			err := dboxpaper.doAPI(
				context.Background(),
				http.MethodPost,
				"https://api.dropboxapi.com/2/paper/docs/list",
				&request{
					out: &buf,
				})
			if err != nil {
				return err
			}
			var docslist DocsList
			err = json.NewDecoder(&buf).Decode(&docslist)
			if err != nil {
				return err
			}
			if c.Bool("l") {
				for _, item := range docslist.DocIds {
					var in bytes.Buffer
					err = json.NewEncoder(&in).Encode(map[string]interface{}{"doc_id": item})
					if err != nil {
						return err
					}
					err = dboxpaper.doAPI(
						context.Background(),
						http.MethodPost,
						"https://api.dropboxapi.com/2/paper/docs/get_metadata",
						&request{
							ct:  "application/json",
							in:  &in,
							out: &buf,
						})
					if err != nil {
						continue
					}
					var docmeta DocsMeta
					err = json.NewDecoder(&buf).Decode(&docmeta)
					if err != nil {
						return err
					}
					fmt.Println(docmeta.DocID, docmeta.Title)
				}
			} else {
				for _, item := range docslist.DocIds {
					fmt.Println(item)
				}
			}
			return nil
		},
	}
	command.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "l",
			Usage: "list title",
		},
	}
	app.Commands = append(app.Commands, command)
}

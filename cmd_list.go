package main

import (
	"context"
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
			var docslist DocsList
			err := dboxpaper.doAPI(context.Background(), http.MethodPost, "https://api.dropboxapi.com/2/paper/docs/list", "", &docslist, nil)
			if err != nil {
				return err
			}
			if c.Bool("l") {
				for _, item := range docslist.DocIds {
					var docmeta DocsMeta
					err = dboxpaper.doAPI(context.Background(), http.MethodPost, "https://api.dropboxapi.com/2/paper/docs/get_metadata", map[string]string{"doc_id": item}, &docmeta, nil)
					if err != nil {
						continue
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

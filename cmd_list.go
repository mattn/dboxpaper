package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/urfave/cli"
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "Show paper items",
		Action: func(c *cli.Context) error {
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			var docslist DocsList
			err := dboxpaper.doAPI(context.Background(), http.MethodPost, "https://api.dropboxapi.com/2/paper/docs/list", "", &docslist)
			if err != nil {
				return err
			}
			for _, item := range docslist.DocIds {
				fmt.Println(item)
			}
			return nil
		},
	})
}

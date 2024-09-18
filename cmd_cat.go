package main

import (
	"context"
	"net/http"

	"github.com/urfave/cli/v2"
)

func init() {
	command := &cli.Command{
		Name:    "cat",
		Aliases: []string{},
		Usage:   "Cat paper",
		Action: func(c *cli.Context) error {
			if !c.Args().Present() {
				cli.ShowCommandHelp(c, "cat")
				return nil
			}
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			return dboxpaper.doAPI(
				context.Background(),
				http.MethodPost,
				"/2/paper/docs/download",
				&request{
					ct:  "application/octet-stream",
					arg: map[string]interface{}{"doc_id": c.Args().First(), "export_format": "markdown"},
					out: c.App.Writer,
				})
		},
	}
	command.BashComplete = docIDCompletion
	app.Commands = append(app.Commands, command)
}

package main

import (
	"context"
	"net/http"
	"os"

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
			return dboxpaper.doAPI(
				context.Background(),
				http.MethodPost,
				"https://api.dropboxapi.com/2/paper/docs/download",
				&request{
					ct:  "application/octet-stream",
					arg: map[string]interface{}{"doc_id": c.Args().First(), "export_format": "markdown"},
					out: os.Stdout,
				})
		},
	}
	command.BashComplete = docIdCompletion
	app.Commands = append(app.Commands, command)
}

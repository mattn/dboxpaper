package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

func init() {
	command := &cli.Command{
		Name:    "delete",
		Aliases: []string{},
		Usage:   "Delete paper permanently",
		Action: func(c *cli.Context) error {
			if !c.Args().Present() {
				cli.ShowCommandHelp(c, "delete")
				return nil
			}
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			return dboxpaper.rpc(
				context.Background(),
				"/2/paper/docs/permanently_delete",
				map[string]interface{}{"doc_id": c.Args().First()},
				nil)
		},
	}
	command.BashComplete = docIDCompletion
	app.Commands = append(app.Commands, command)
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/urfave/cli"
)

func init() {
	command := cli.Command{
		Name:    "delete",
		Aliases: []string{},
		Usage:   "Delete paper permanently",
		Action: func(c *cli.Context) error {
			if !c.Args().Present() {
				cli.ShowCommandHelp(c, "delete")
				return nil
			}
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			var in bytes.Buffer
			err := json.NewEncoder(&in).Encode(map[string]interface{}{"doc_id": c.Args().First()})
			if err != nil {
				return err
			}
			return dboxpaper.doAPI(
				context.Background(),
				http.MethodPost,
				"/2/paper/docs/permanently_delete",
				&request{
					ct: "application/json",
					in: &in,
				})
		},
	}
	command.BashComplete = docIDCompletion
	app.Commands = append(app.Commands, command)
}

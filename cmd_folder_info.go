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
		Name:    "folder_info",
		Aliases: []string{},
		Usage:   "Show folder information for the paper",
		Action: func(c *cli.Context) error {
			if !c.Args().Present() {
				cli.ShowCommandHelp(c, "folder_info")
				return nil
			}
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			var in, out bytes.Buffer
			err := json.NewEncoder(&in).Encode(map[string]interface{}{"doc_id": c.Args().First()})
			if err != nil {
				return err
			}
			err = dboxpaper.doAPI(
				context.Background(),
				http.MethodPost,
				"/2/paper/docs/get_folder_info",
				&request{
					ct:  "application/json",
					in:  &in,
					out: &out,
				})
			if err != nil {
				return err
			}
			var folderinfo FolderInfo
			err = json.NewDecoder(&out).Decode(&folderinfo)
			if err != nil {
				return err
			}
			for _, folder := range folderinfo.Folders {
				fmt.Fprintf(c.App.Writer, "%s %s\n", folder.ID, folder.Name)
			}
			return nil
		},
	}
	command.BashComplete = docIDCompletion
	app.Commands = append(app.Commands, command)
}

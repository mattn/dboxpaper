package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
)

func init() {
	command := &cli.Command{
		Name:      "folder_info",
		Aliases:   []string{},
		Usage:     "Show folder information for the paper",
		ArgsUsage: "<doc_id>",
		Action: func(c *cli.Context) error {
			if !c.Args().Present() {
				cli.ShowCommandHelp(c, "folder_info")
				return nil
			}
			if isFilePath(c.Args().First()) {
				return errors.New("folder_info is not supported for .paper files")
			}
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			var folderinfo FolderInfo
			err := dboxpaper.rpc(
				context.Background(),
				"/2/paper/docs/get_folder_info",
				map[string]interface{}{"doc_id": c.Args().First()},
				&folderinfo)
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

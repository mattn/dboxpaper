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
			docIds, err := listDocs(c)
			if !c.Bool("l") {
				for _, item := range docIds {
					fmt.Fprintln(c.App.Writer, item)
				}
				return nil
			}
			for _, item := range docIds {
				var in, out bytes.Buffer
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
						out: &out,
					})
				if err != nil {
					continue
				}
				var docmeta DocsMeta
				err = json.NewDecoder(&out).Decode(&docmeta)
				if err != nil {
					return err
				}
				fmt.Println(docmeta.DocID, docmeta.Title)
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

func listDocs(c *cli.Context) ([]string, error) {
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
		return nil, err
	}
	var docslist DocsList
	err = json.NewDecoder(&buf).Decode(&docslist)
	if err != nil {
		return nil, err
	}
	return docslist.DocIds, nil
}

func docIdCompletion(c *cli.Context) {
	docIds, err := listDocs(c)
	if err != nil {
		return
	}
	for _, item := range docIds {
		fmt.Fprintln(c.App.Writer, item)
	}
}

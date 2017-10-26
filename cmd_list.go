package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
			if c.Bool("title") {
				for _, item := range docIds {
					var in, out bytes.Buffer
					err = json.NewEncoder(&in).Encode(map[string]interface{}{"doc_id": item})
					if err != nil {
						return err
					}
					err = dboxpaper.doAPI(
						context.Background(),
						http.MethodPost,
						"/2/paper/docs/get_metadata",
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
					fmt.Fprintf(c.App.Writer, "%s %s\n", docmeta.DocID, docmeta.Title)
				}
			} else if c.Bool("json") {
				fmt.Fprint(c.App.Writer, "[")
				n := 0
				for _, item := range docIds {
					var in, out bytes.Buffer
					err = json.NewEncoder(&in).Encode(map[string]interface{}{"doc_id": item})
					if err != nil {
						return err
					}
					if dboxpaper.doAPI(
						context.Background(),
						http.MethodPost,
						"/2/paper/docs/get_metadata",
						&request{
							ct:  "application/json",
							in:  &in,
							out: &out,
						}) != nil {
						continue
					}
					if n > 0 {
						fmt.Fprint(c.App.Writer, ",")
					}
					n++
					io.Copy(c.App.Writer, &out)
				}
				fmt.Fprint(c.App.Writer, "]")
			} else {
				for _, item := range docIds {
					fmt.Fprintln(c.App.Writer, item)
				}
				return nil
			}
			return nil
		},
	}
	command.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "title",
			Usage: "show title",
		},
		cli.BoolFlag{
			Name:  "json",
			Usage: "show as JSON",
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
		"/2/paper/docs/list",
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

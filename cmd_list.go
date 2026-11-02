package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
)

func init() {
	command := &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "Show papers",
		Action: func(c *cli.Context) error {
			dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
			docIds, err := listDocs(c)
			if err != nil {
				return err
			}
			if !c.Bool("title") && !c.Bool("json") {
				for _, item := range docIds {
					fmt.Fprintln(c.App.Writer, item)
				}
				return nil
			}

			metas := []json.RawMessage{}
			for _, item := range docIds {
				docmeta, raw, err := dboxpaper.getMetadata(context.Background(), item)
				if err != nil {
					fmt.Fprintf(c.App.ErrWriter, "%s: %v\n", item, err)
					continue
				}
				if c.Bool("json") {
					metas = append(metas, raw)
				} else {
					fmt.Fprintf(c.App.Writer, "%s %s\n", docmeta.DocID, docmeta.Title)
				}
			}
			if c.Bool("json") {
				return json.NewEncoder(c.App.Writer).Encode(metas)
			}
			return nil
		},
	}
	command.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "title",
			Usage: "show title",
		},
		&cli.BoolFlag{
			Name:  "json",
			Usage: "show as JSON",
		},
	}
	app.Commands = append(app.Commands, command)
}

// rpc calls an RPC style endpoint with JSON argument and decodes JSON result into v.
func (dboxpaper *DboxPaper) rpc(ctx context.Context, path string, arg interface{}, v interface{}) error {
	req := &request{}
	if arg != nil {
		var in bytes.Buffer
		if err := json.NewEncoder(&in).Encode(arg); err != nil {
			return err
		}
		req.ct = "application/json"
		req.in = &in
	}
	var out bytes.Buffer
	req.out = &out
	if err := dboxpaper.doAPI(ctx, http.MethodPost, path, req); err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	return json.NewDecoder(&out).Decode(v)
}

func (dboxpaper *DboxPaper) getMetadata(ctx context.Context, docID string) (*DocsMeta, json.RawMessage, error) {
	var raw json.RawMessage
	err := dboxpaper.rpc(ctx, "/2/paper/docs/get_metadata", map[string]interface{}{"doc_id": docID}, &raw)
	if err != nil {
		return nil, nil, err
	}
	var docmeta DocsMeta
	if err := json.Unmarshal(raw, &docmeta); err != nil {
		return nil, nil, err
	}
	return &docmeta, raw, nil
}

func listDocs(c *cli.Context) ([]string, error) {
	dboxpaper := app.Metadata["dboxpaper"].(*DboxPaper)
	var docIds []string
	var docslist DocsList
	err := dboxpaper.rpc(context.Background(), "/2/paper/docs/list", nil, &docslist)
	for {
		if err != nil {
			return nil, err
		}
		docIds = append(docIds, docslist.DocIds...)
		if !docslist.HasMore {
			return docIds, nil
		}
		cursor := docslist.Cursor.Value
		docslist = DocsList{}
		err = dboxpaper.rpc(context.Background(), "/2/paper/docs/list/continue", map[string]interface{}{"cursor": cursor}, &docslist)
	}
}

func docIDCompletion(c *cli.Context) {
	// Do not start the OAuth flow from shell completion.
	if app.Metadata["dboxpaper"].(*DboxPaper).needsLogin() {
		return
	}
	docIds, err := listDocs(c)
	if err != nil {
		return
	}
	for _, item := range docIds {
		fmt.Fprintln(c.App.Writer, item)
	}
}

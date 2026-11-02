package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/urfave/cli/v2"
)

func testWithServer(h http.HandlerFunc, testFuncs ...func(*cli.App)) string {
	ts := httptest.NewServer(h)
	defer ts.Close()

	cli.OsExiter = func(n int) {}

	uri, _ := url.Parse(ts.URL)

	var buf bytes.Buffer
	app.Metadata = map[string]interface{}{
		"dboxpaper": &DboxPaper{
			uri:        uri,
			contentURI: uri,
		},
	}
	app.Writer = &buf

	for _, f := range testFuncs {
		f(app)
	}

	return buf.String()
}

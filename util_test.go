package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"

	"golang.org/x/oauth2"

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
			uri:   uri,
			token: nil,
			config: &oauth2.Config{
				Scopes: []string{},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://www.dropbox.com/oauth2/authorize",
					TokenURL: "https://api.dropboxapi.com/oauth2/token",
				},
				ClientID:     "nrb8y95k7yoeor6",
				ClientSecret: "fhme6tzwkzw5og8",
				RedirectURL:  "http://localhost:8989",
			},
		},
	}
	app.Writer = &buf

	for _, f := range testFuncs {
		f(app)
	}

	return buf.String()
}

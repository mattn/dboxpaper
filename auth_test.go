package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestTokenRefresh(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"access_token":"new","token_type":"bearer","expires_in":14400}`)
		case "/2/paper/docs/list":
			if got := r.Header.Get("Authorization"); got != "Bearer new" {
				t.Errorf("want refreshed token but got %q", got)
			}
			fmt.Fprintln(w, `{"doc_ids":[]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	uri, _ := url.Parse(ts.URL)
	file := filepath.Join(t.TempDir(), "settings.json")
	d := &DboxPaper{
		uri:  uri,
		file: file,
		token: &oauth2.Token{
			AccessToken:  "old",
			RefreshToken: "refresh",
			Expiry:       time.Now().Add(-time.Hour),
		},
		config: &oauth2.Config{
			Endpoint: oauth2.Endpoint{TokenURL: ts.URL + "/oauth2/token"},
		},
	}
	if d.needsLogin() {
		t.Fatal("token with refresh token should not need login")
	}
	if err := d.doAPI(context.Background(), http.MethodPost, "/2/paper/docs/list", &request{}); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var saved oauth2.Token
	if err := json.Unmarshal(b, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.AccessToken != "new" || saved.RefreshToken != "refresh" {
		t.Fatalf("unexpected saved token: %+v", saved)
	}
}

func TestNeedsLogin(t *testing.T) {
	cfg := &oauth2.Config{}
	tests := []struct {
		token *oauth2.Token
		want  bool
	}{
		{nil, true},
		{&oauth2.Token{AccessToken: "x"}, false},
		{&oauth2.Token{AccessToken: "x", Expiry: time.Now().Add(-time.Hour)}, true},
		{&oauth2.Token{AccessToken: "x", RefreshToken: "r", Expiry: time.Now().Add(-time.Hour)}, false},
	}
	for i, tt := range tests {
		d := &DboxPaper{config: cfg, token: tt.token}
		if got := d.needsLogin(); got != tt.want {
			t.Errorf("#%d: want %v but got %v", i, tt.want, got)
		}
	}
}

package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"

	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli/v2"
)

const name = "dboxpaper"

const version = "0.0.2"

var revision = "HEAD"

// DocsMeta is struct for dropbox paper API
type DocsMeta struct {
	DocID       string    `json:"doc_id"`
	Owner       string    `json:"owner"`
	Title       string    `json:"title"`
	CreatedDate time.Time `json:"created_date"`
	Status      struct {
		Tag string `json:".tag"`
	} `json:"status"`
	Revision        int       `json:"revision"`
	LastUpdatedDate time.Time `json:"last_updated_date"`
	LastEditor      string    `json:"last_editor"`
}

// DocsList is struct for dropbox paper API
type DocsList struct {
	DocIds []string `json:"doc_ids"`
	Cursor struct {
		Value      string    `json:"value"`
		Expiration time.Time `json:"expiration"`
	} `json:"cursor"`
	HasMore bool `json:"has_more"`
}

// FolderInfo is struct for dropbox paper API
type FolderInfo struct {
	FolderSharingPolicyType struct {
		Tag string `json:".tag"`
	} `json:"folder_sharing_policy_type"`
	Folders []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"folders"`
}

var (
	app = cli.NewApp()
)

// DboxPaper is application for dboxpaper
type DboxPaper struct {
	uri    *url.URL
	token  *oauth2.Token
	config *oauth2.Config
	file   string
	debug  bool
}

type request struct {
	ct  string
	arg map[string]interface{}
	in  io.Reader
	out io.Writer
}

func (dboxpaper *DboxPaper) doAPI(ctx context.Context, method string, path string, request *request) error {
	if err := dboxpaper.authorize(ctx); err != nil {
		return err
	}
	u := *dboxpaper.uri
	u.Path = path
	req, err := http.NewRequestWithContext(ctx, method, u.String(), request.in)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", request.ct)
	if request.arg != nil {
		b, err := json.Marshal(request.arg)
		if err != nil {
			return err
		}
		req.Header.Add("Dropbox-API-Arg", string(b))
	}
	resp, err := dboxpaper.httpClient(ctx).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var r io.Reader = resp.Body

	if dboxpaper.debug {
		r = io.TeeReader(r, os.Stderr)
	}

	if resp.StatusCode >= 400 {
		b, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		if len(b) == 0 {
			return errors.New(resp.Status)
		}
		return errors.New(string(b))
	}

	if request.out != nil {
		_, err = io.Copy(request.out, r)
	} else {
		_, err = io.Copy(io.Discard, r)
	}
	return err
}

// needsLogin reports whether the OAuth flow must be run before calling the API.
func (dboxpaper *DboxPaper) needsLogin() bool {
	if dboxpaper.config == nil {
		return false
	}
	t := dboxpaper.token
	return t == nil || (!t.Valid() && t.RefreshToken == "")
}

func (dboxpaper *DboxPaper) authorize(ctx context.Context) error {
	if !dboxpaper.needsLogin() {
		return nil
	}
	if err := dboxpaper.AccessToken(ctx); err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}
	return nil
}

func (dboxpaper *DboxPaper) httpClient(ctx context.Context) *http.Client {
	if dboxpaper.config == nil || dboxpaper.token == nil {
		return http.DefaultClient
	}
	return oauth2.NewClient(ctx, &savingTokenSource{
		dboxpaper: dboxpaper,
		src:       dboxpaper.config.TokenSource(ctx, dboxpaper.token),
	})
}

// savingTokenSource stores the token to the settings file when it is refreshed.
type savingTokenSource struct {
	dboxpaper *DboxPaper
	src       oauth2.TokenSource
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	t, err := s.src.Token()
	if err != nil {
		return nil, err
	}
	if t.AccessToken != s.dboxpaper.token.AccessToken {
		s.dboxpaper.token = t
		if err := s.dboxpaper.saveToken(); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func configDir() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", name), nil
	}
	if runtime.GOOS == "windows" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, name), nil
	}
	return "", errors.New("$HOME is not defined")
}

// Setup is API to setup application
func (dboxpaper *DboxPaper) Setup() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	dboxpaper.file = filepath.Join(dir, "settings.json")

	b, err := os.ReadFile(dboxpaper.file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &dboxpaper.token)
	if err != nil {
		return fmt.Errorf("could not unmarshal %v: %v", dboxpaper.file, err)
	}
	return nil
}

func (dboxpaper *DboxPaper) saveToken() error {
	b, err := json.MarshalIndent(dboxpaper.token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to store file: %v", err)
	}
	if err := os.WriteFile(dboxpaper.file, b, 0600); err != nil {
		return fmt.Errorf("failed to store file: %v", err)
	}
	// WriteFile does not change the mode of an existing file.
	if err := os.Chmod(dboxpaper.file, 0600); err != nil {
		return fmt.Errorf("failed to store file: %v", err)
	}
	return nil
}

// AccessToken is API to get access token
func (dboxpaper *DboxPaper) AccessToken(ctx context.Context) error {
	l, err := net.Listen("tcp", "localhost:8989")
	if err != nil {
		return err
	}

	stateBytes := make([]byte, 16)
	_, err = rand.Read(stateBytes)
	if err != nil {
		return err
	}
	state := fmt.Sprintf("%x", stateBytes)
	verifier := oauth2.GenerateVerifier()

	uri := dboxpaper.config.AuthCodeURL(state,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("token_access_type", "offline"))
	fmt.Fprintf(os.Stderr, "opening browser: %s\n", uri)
	if err := open.Start(uri); err != nil {
		fmt.Fprintf(os.Stderr, "failed to open browser, please open the URL above manually: %v\n", err)
	}

	type result struct {
		code string
		err  error
	}
	ch := make(chan result, 1)
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			q := req.URL.Query()
			if req.URL.Path != "/" || q.Get("state") != state {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			var res result
			if e := q.Get("error"); e != "" {
				res.err = fmt.Errorf("%s: %s", e, q.Get("error_description"))
			} else if res.code = q.Get("code"); res.code == "" {
				res.err = errors.New("authorization code is missing")
			}
			if res.err != nil {
				http.Error(w, res.err.Error(), http.StatusBadRequest)
			} else {
				fmt.Fprintln(w, "Authorized. You can close this window.")
			}
			select {
			case ch <- res:
			default:
			}
		}),
	}
	go srv.Serve(l)
	defer srv.Close()

	var res result
	select {
	case res = <-ch:
	case <-ctx.Done():
		return ctx.Err()
	}
	if res.err != nil {
		return res.err
	}

	token, err := dboxpaper.config.Exchange(ctx, res.code, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("failed to exchange access-token: %v", err)
	}
	dboxpaper.token = token
	return dboxpaper.saveToken()
}

func initialize(c *cli.Context) error {
	dboxpaper := &DboxPaper{
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
	}
	err := dboxpaper.Setup()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to get configuration: %v", err)
	}

	dboxpaper.uri, _ = url.Parse("https://api.dropboxapi.com")

	dboxpaper.debug = os.Getenv("DBOXPAPER_DEBUG") != ""
	app.Metadata["dboxpaper"] = dboxpaper
	app.Metadata["stdin"] = os.Stdin
	return nil
}

func main() {
	app.Name = "dboxpaper"
	app.Usage = "Dropbox Paper client"
	app.Version = version
	app.Authors = []*cli.Author{
		{
			Name:  "mattn",
			Email: "mattn.jp@gmail.com",
		},
	}
	app.EnableBashCompletion = true
	app.Before = initialize
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

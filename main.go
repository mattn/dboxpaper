package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"

	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
)

type DocsList struct {
	DocIds []string `json:"doc_ids"`
	Cursor struct {
		Value      string    `json:"value"`
		Expiration time.Time `json:"expiration"`
	} `json:"cursor"`
	HasMore bool `json:"has_more"`
}

type config map[string]string

var (
	app = cli.NewApp()
)

func init() {
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "json",
			Usage: "output json",
		},
	}
}

type DboxPaper struct {
	token  *oauth2.Token
	config *oauth2.Config
	file   string
}

func (dboxpaper *DboxPaper) doAPI(ctx context.Context, method string, uri string, params interface{}, res interface{}) error {
	var stream io.Reader
	var args string
	if params != nil {
		if qargs, ok := params.(string); ok {
			buf := new(bytes.Buffer)
			buf.WriteString("{}")
			args = qargs
		} else {
			buf := new(bytes.Buffer)
			err := json.NewEncoder(buf).Encode(params)
			if err != nil {
				return err
			}
			stream = buf
		}
	}

	req, err := http.NewRequest(method, uri, stream)
	if err != nil {
		return err
	}
	req.WithContext(ctx)
	req.Header.Add("Authorization", "Bearer "+dboxpaper.token.AccessToken)
	if args != "" {
		req.Header.Add("Dropbox-API-Arg", args)
	}
	client := dboxpaper.config.Client(ctx, dboxpaper.token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var r io.Reader = resp.Body
	//r = io.TeeReader(resp.Body, os.Stdout)

	if res != nil {
		if s, ok := res.(*string); ok {
			b, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				*s = string(b)
			}
		} else {
			err = json.NewDecoder(r).Decode(res)
		}
	} else {
		_, err = io.Copy(ioutil.Discard, r)
	}
	return err
}

func (dboxpaper *DboxPaper) Setup() error {
	dir := os.Getenv("HOME")
	if dir == "" && runtime.GOOS == "windows" {
		dir = os.Getenv("APPDATA")
		if dir == "" {
			dir = filepath.Join(os.Getenv("USERPROFILE"), "Application Data", "dboxpaper")
		}
		dir = filepath.Join(dir, "dboxpaper")
	} else {
		dir = filepath.Join(dir, ".config", "dboxpaper")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	dboxpaper.file = filepath.Join(dir, "settings.json")

	b, err := ioutil.ReadFile(dboxpaper.file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &dboxpaper.token)
	if err != nil {
		return fmt.Errorf("could not unmarshal %v: %v", dboxpaper.file, err)
	}
	return nil
}

func (dboxpaper *DboxPaper) AccessToken() error {
	l, err := net.Listen("tcp", "localhost:8989")
	if err != nil {
		return err
	}
	defer l.Close()

	stateBytes := make([]byte, 16)
	_, err = rand.Read(stateBytes)
	if err != nil {
		return err
	}

	state := fmt.Sprintf("%x", stateBytes)

	err = open.Start(dboxpaper.config.AuthCodeURL(state, oauth2.SetAuthURLParam("response_type", "code")))
	if err != nil {
		return err
	}

	quit := make(chan string)
	go http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		code := req.URL.Query().Get("code")
		if code == "" {
			w.Write([]byte(`<script>document.write(location.hash)</script>`))
		} else {
			w.Write([]byte(`<script>window.open("about:blank","_self").close()</script>`))
		}
		w.(http.Flusher).Flush()
		quit <- code
	}))

	dboxpaper.token, err = dboxpaper.config.Exchange(context.Background(), <-quit)
	if err != nil {
		return fmt.Errorf("failed to exchange access-token: %v", err)
	}

	b, err := json.MarshalIndent(dboxpaper.token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to store file: %v", err)
	}
	err = ioutil.WriteFile(dboxpaper.file, b, 0700)
	if err != nil {
		return fmt.Errorf("failed to store file: %v", err)
	}
	return nil
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

	if dboxpaper.token == nil || dboxpaper.token.AccessToken == "" {
		err = dboxpaper.AccessToken()
		if err != nil {
			return fmt.Errorf("faild to get access token: %v", err)
		}
	}

	app.Metadata["dboxpaper"] = dboxpaper
	return nil
}

func main() {
	app.Name = "dboxpaper"
	app.Usage = "Dropbox Paper client"
	app.Version = "0.0.1"
	app.Before = initialize
	app.Setup()
	app.Run(os.Args)
}

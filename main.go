package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHTTP "github.com/go-git/go-git/v5/plumbing/transport/http"
	"log"
)

// App struct
type App struct {
	serverDomain string
	repo         *git.Repository
	auth         transport.AuthMethod
	config       *Config
}

// assume run in project directory containing .git folder
// perform git pull
// abort if merge conflict
func main() {
	initFlags()
	cfg := newConfig()
	app := NewApp(cfg)
	app.startup()
	app.PullWithAbort()
}

// NewApp creates a new App application struct
func NewApp(cfg *Config) *App {
	a := &App{}
	a.config = cfg
	return a
}

// startup is called when the app starts.
func (a *App) startup() {
	// print config
	fmt.Printf("%+v\n", a.config)

	// create basic auth with username and personal access token
	a.auth = &gitHTTP.BasicAuth{
		Username: a.config.Username,
		Password: a.config.PersonalAccessToken,
	}

	// open repository
	if err := a.OpenRepository(a.config.RepoDirectory); err != nil {
		log.Fatalf("could not open repository: %s\n", err)
	}
}

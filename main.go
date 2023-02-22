package main

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"io"
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
	app := NewApp()
	app.startup()
	app.PullWithAbort()
}

// NewApp creates a new App application struct
func NewApp() *App {
	a := &App{}
	a.config = &Config{}
	return a
}

// startup is called when the app starts.
func (a *App) startup() {
	a.serverDomain = "https://git.klei.com"

	// read config file
	if err := a.ReadConfig(); err != nil {
		if errors.Is(err, io.EOF) {
			log.Printf("empty config file: %s", err)
		}
		fmt.Printf("read config: %s\n", err)
		if !errors.Is(err, io.EOF) {
			log.Fatal(err)
		}
	}
	fmt.Printf("Username: %s, RepoDirectory: %s\n", a.config.Username, a.config.RepoDirectory)
	fmt.Printf("Use read config? y/n\n")
	var input string
	fmt.Scanln(&input)
	switch input {
	case "n":
		a.config.RepoDirectory = RepoDirectory()
		username, password, err := credentials()
		if err != nil {
			log.Fatal(err)
		}
		a.config.Username = username
		a.config.PersonalAccessToken = password
	case "y":
	default:
	}
	if err := a.OpenRepository(a.config.RepoDirectory); err != nil {
		log.Fatalf("could not open repository: %s\n", err)
	}
	if err := a.BasicAuth(); err != nil {
		log.Fatal(err)
	}

	err := a.WriteConfig()
	if err != nil {
		log.Fatal(err)
	}
}

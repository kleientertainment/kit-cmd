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
// upon user input, perform git pull
// abort if merge conflict
func main() {
	app := NewApp()
	app.startup()

}

// NewApp creates a new App application struct
func NewApp() *App {
	a := &App{}
	a.serverDomain = "https://git.klei.com"
	a.config = &Config{}
	a.config.WorkingDirectory = ExecutableDirectory()
	return a
}

// startup is called when the app starts.
func (a *App) startup() {
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

	if err := a.BasicAuth(); err != nil {
		log.Fatal(err)
	}

	if err := a.OpenRepository(a.config.WorkingDirectory); err != nil {
		fmt.Printf("open repository: %s\n", err)
	}
}

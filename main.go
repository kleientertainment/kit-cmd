package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHTTP "github.com/go-git/go-git/v5/plumbing/transport/http"
	"log"
	"os/exec"
)

// App struct
type App struct {
	serverDomain string
	repo         *git.Repository
	auth         transport.AuthMethod
	config       *Config
}

var app *App

func initializeApplication() {
	initFlags()
	app = &App{}
	app.config = newConfig()
	app.startup()
}

// startup is called when the app starts.
func (a *App) startup() {
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

// this is a pull script
func main() {
	initializeApplication()

	err := ExecPull()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func ExecPull() error {
	var err error
	var cmd *exec.Cmd

	cmd = exec.Command("git", "pull", "--rebase=false") // --rebase flag specify for system-specific config
	if err = cmdWrapperPrintOutput(cmd, app.config.RepoDirectory); err != nil {
		if mErr := ExecAbortMerge(); mErr != nil {
			return fmt.Errorf("pull error. could not abort merge")
		}
		return fmt.Errorf("pull error. merge aborted")
	}
	fmt.Printf("Pull successful!\n")
	return nil
}

func ExecAbortMerge() error {
	cmd := exec.Command("git", "merge", "--abort")
	if err := cmdWrapperPrintOutput(cmd, app.config.RepoDirectory); err != nil {
		return fmt.Errorf("merge abort error: %s\n", err)
	}
	return nil
}

func cmdWrapperPrintOutput(cmd *exec.Cmd, dir string) error {
	cmd.Dir = dir
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	// Print the output
	fmt.Println(string(stdout))
	return nil
}

func Alert(e error, method string) {
	fmt.Printf("Get a programmer to help with this:\n%s error: %s\n", method, e)
}

//
//err = app.Push()
//if err != nil {
//	Alert(err, "Push")
//}

package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHTTP "github.com/go-git/go-git/v5/plumbing/transport/http"
	"log"
	"os/exec"
	"strings"
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
	var err error
	initializeApplication()

	if err = PullCmd("--ff-only"); err != nil {
		if err = PullCmd("--no-rebase"); err != nil {
			fmt.Printf("Could not pull. Attempting to abort %s\n", err)
			if err = AbortMerge(); err != nil {
				fmt.Printf("%s\n", err)
			}
		}
		//err = AbortMerge()
	}

	status, err := app.StatusCmd()
	if err != nil {
		log.Fatalf("status error: %s\n", err)
	}
	fmt.Printf("%s\n", status)

	//lockers := []string{"WindowsFileLocker.exe", "photoshop.exe", "animate.exe", "gameXYZ.exe"}
	//lockFlag := false
	//for _, s := range lockers {
	//	running, err := imageNameRunning(s)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	if running == true {
	//		lockFlag = true
	//		fmt.Printf("Potentially locking %s open, please close\n", s)
	//	}
	//}

	//if lockFlag == false { // proceed
	//}
}

//func imageNamePIDs(imageName string) ([]int, error) {
//	pids := make([]int, 5)
//	filterArg := "IMAGENAME eq " + imageName
//	cmd := exec.Command("tasklist", "/fi", filterArg)
//	stdout, err := cmd.Output()
//	if err != nil {
//		return nil, err
//	}
//	//out := string(stdout)
//	// parse out for pid, add to pids
//
//	return pids, nil
//}

// imageName: animate.exe , photoshop.exe, gameXYZ.exe, WindowsFileLocker.exe
func imageNameRunning(imageName string) (bool, error) {
	filterArg := "IMAGENAME eq " + imageName
	cmd := exec.Command("tasklist", "/fi", filterArg)
	stdout, err := cmd.Output()
	if err != nil {
		return false, err
	}
	out := string(stdout)
	if strings.Contains(out, imageName) {
		fmt.Printf("%s\n", out)
		return true, nil
	}
	return false, nil
}

func Alert(e error) {
	fmt.Printf("Get a programmer to help with this:\n%s\n", e)
}

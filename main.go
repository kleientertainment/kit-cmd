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
// requires git configured on command line with creds for the repo
func main() {
	var err error
	initializeApplication()

	output, err := PullCmd("--ff-only")
	if err != nil {
		fmt.Printf("%s\n", output)
		output, err = PullCmd("--no-rebase")

		if err != nil {
			fmt.Printf("Get a programmer to help with this pull!\n-----------------------\nCould not pull. Attempting cleanup...\n")
			if err = AbortMerge(); err != nil {
				fmt.Printf("Merge abort error: %s\n", err)
			}
		}
	}

	status, err := app.StatusCmd()
	if err != nil {
		log.Fatalf("status error: %s\n", err)
	}
	fmt.Printf("%s\n", status)

	fmt.Printf("-----------------------\nSearching for locking programs...\n")
	lockers := []string{"WindowsFileLocker.exe", "Photoshop.exe", "Animate.exe", "GLN_r.exe", "GLN.exe"}
	//lockFlag := false
	for _, s := range lockers {
		running, err := imageNameRunning(s)
		if err != nil {
			log.Fatal(err)
		}
		if running == true {
			//lockFlag = true
			fmt.Printf("Potentially locking %s open, consider closing\n", s)
		}
	}

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
		return true, nil
	}
	return false, nil
}

func Alert(e error) {
	fmt.Printf("Get a programmer to help with this:\n%s\n", e)
}

package main

import (
	"fmt"
	"github.com/libgit2/git2go/v34"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"pkg.klei.ca/kit_cmd/internal"
	"strings"
)

// App struct
type App struct {
	serverDomain string
	repo         *git.Repository
	//auth         transport.AuthMethod
	config *internal.Config
}

var app *App

func initializeApplication() {
	internal.initFlags()
	app = &App{}
	app.config = internal.newConfig()
	app.startup()
}

// startup is called when the app starts.
func (a *App) startup() {
	// create basic auth with username and personal access token
	//a.auth = &gitHTTP.BasicAuth{
	//	Username: a.config.Username,
	//	Password: a.config.PersonalAccessToken,
	//}

	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 3; i++ { // gets us to root dir of project "GLN/tools/git_scripts/pull" -> "GLN"
		path = filepath.Dir(path)
	}
	app.config.RepoDirectory = path

	// open repository
	repo, err := git.OpenRepository(a.config.RepoDirectory)
	if err != nil {
		log.Fatalf("could not open repository: %s\n", err)
	}
	a.repo = repo
}

// this is a pull script
func main() {
	//var err error
	initializeApplication()

	/*
		output, err := PullCmd("--ff-only")
		if err != nil {
			fmt.Printf("%s\n", output)
			output, err = PullCmd("--no-rebase")

			if err != nil {
				fmt.Printf("Get a programmer to help with this pull!\n-----------------------\nCould not pull. Attempting cleanup...\n")
				if err = AbortMerge(); err != nil {
					fmt.Printf("Merge abort error: %s\n", err)
				} else {
					fmt.Printf("Cleaned!\n\n")
				}

				fmt.Printf("-----------------------\nSearching for locking programs...\n")
				lockers := []string{"WindowsFileLocker.exe", "Photoshop.exe", "Animate.exe", "GLN_r.exe", "GLN.exe"}
				//lockFlag := false
				for _, s := range lockers {
					running, err := imageNameRunning(s)
					if err != nil {
						fmt.Printf("%s\n", err)
					}
					if running == true {
						//lockFlag = true
						fmt.Printf("Potentially locking %s open, watch out for file locks!\n", s)
					}
				}
			}
		}

		status, err := app.StatusCmd()
		if err != nil {
			log.Fatalf("status error: %s\n", err)
		}
		fmt.Printf("%s\n", status)
	*/
}

func exit() {
	fmt.Scanf("g") // hacky way to keep the console window open after execution finishes
}

func imageNamePIDs(imageName string) ([]int, error) {
	pids := make([]int, 5)
	filterArg := "IMAGENAME eq " + imageName
	cmd := exec.Command("tasklist", "/fi", filterArg)
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	out := string(stdout)
	// parse out for pid, add to pids
	fmt.Printf("%s\n", out)
	return pids, nil
}

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

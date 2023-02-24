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
	initializeApplication()

	//status, err := app.Status()
	//if err != nil {
	//	log.Fatalf("status error: %s\n", err)
	//}
	//fmt.Printf("%s\n", status.String())
	//for k, v := range *status {
	//	fmt.Printf("%s %+v\n", k, *v)
	//}

	err := ExecPull()
	if err != nil {
		Alert(err)
	}
	//	status, err := app.Status()
	//	if err != nil {
	//		log.Fatalf("status error: %s\n", err)
	//	}
	//	fmt.Printf("%s\n", status.String())
	//}

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

func ExecPull() error {
	var err error
	var cmd *exec.Cmd

	cmd = exec.Command("cmd", "/C", "git", "pull", "--no-rebase") // same as --rebase=false. specifies to use merge commit
	if err = cmdWrapperPrintOutput(cmd, app.config.RepoDirectory); err != nil {
		if mErr := ExecAbortMerge(); mErr != nil {
			return fmt.Errorf("pull error. could not abort merge: %s: %s\n", err, mErr)
		}
		return fmt.Errorf("pull error. merge aborted after initiation: %s\n", err)
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

func Alert(e error) {
	fmt.Printf("Get a programmer to help with this:\n%s\n", e)
}

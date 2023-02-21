package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHTTP "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	repo         *git.Repository
	auth         transport.AuthMethod
	config       *Config
	serverDomain string
}

type Config struct {
	WorkingDirectory    string
	PersonalAccessToken string
	Username            string
	Email               string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.serverDomain = "https://git.klei.com"

	a.config = &Config{}
	if err := a.ReadConfig(); err != nil {
		if !errors.Is(err, io.EOF) {
			log.Fatal(err)
		}
	}

	if err := a.BasicAuth(); err != nil {
		log.Fatal(err)
	}

	if a.config.WorkingDirectory != "" {
		err := a.OpenRepository()
		if err != nil && !errors.Is(err, git.ErrRepositoryNotExists) {
			log.Fatal(err)
		}
	}
}

func (a *App) BasicAuth() error {
	a.auth = &gitHTTP.BasicAuth{
		Username: a.config.Username,
		Password: a.config.PersonalAccessToken,
	}
	return nil
}

func (a *App) startBackgroundStatus() {
	ticker := time.NewTicker(time.Second * 5)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				local, err := a.LocalModifiedFiles()
				if err != nil {
					log.Fatal(err)
				}
				runtime.EventsEmit(a.ctx, "backgroundStatus", MapToSlice(local))
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (a *App) startBackgroundConflictDetection() {
	ticker := time.NewTicker(time.Second * 5)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				local, err := a.LocalModifiedFiles()
				if err != nil {
					log.Printf("%s\n", err)
				}
				remote, err := a.RemoteModifiedFiles()
				if err != nil {
					log.Printf("%s\n", err)
				}
				conflicts := a.Conflicts(local, remote)
				if len(conflicts) > 0 {
					runtime.EventsEmit(a.ctx, "backgroundConflictDetection", conflicts)
					err := beeep.Notify("Title", "Message body", "assets/information.png")
					if err != nil {
						panic(err)
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (a *App) startBackgroundFetch() {
	ticker := time.NewTicker(time.Second * 20)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				err := a.Fetch()
				if !errors.Is(err, git.NoErrAlreadyUpToDate) {
					log.Fatal(err)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (a *App) domReady(ctx context.Context) {
	a.ctx = ctx

}

func (a *App) beforeClose(ctx context.Context) bool {
	dialog, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:    runtime.WarningDialog,
		Title:   "Quit?",
		Message: "Are you sure you want to quit?",
		Buttons: []string{"Cancel", "Yes"},
	})
	if err != nil {
		return false
	}
	return dialog != "Yes"
}

func (a *App) shutdown(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SetWorkingDirectory() error {
	defaultDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		DefaultDirectory: defaultDir,
	})
	if err != nil {
		log.Fatal(err)
	}

	a.config.WorkingDirectory = dir
	a.WriteConfig()
	err = a.OpenRepository()
	if err != nil {
		return err
	}
	return nil
}

func (a *App) Pull() {
	w, err := a.repo.Worktree()
	if err != nil {
		log.Fatal(err)
	}
	err = w.Pull(&git.PullOptions{
		Auth: a.auth,
	})
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return
		}
		log.Fatal(err)
	}
}

func (a *App) LocalModifiedFiles() (map[string]struct{}, error) {
	status, err := a.Status()
	if err != nil {
		return nil, err
	}

	files := make(map[string]struct{})
	for k := range *status {
		files[k] = struct{}{}
	}
	return files, nil
}

func MapToSlice(m map[string]struct{}) (s []string) {
	for k := range m {
		s = append(s, k)
	}
	return s
}

func (a *App) Conflicts(local map[string]struct{}, remote map[string]struct{}) []string {
	var conflicts []string
	for k := range local {
		_, ok := remote[k]
		if ok {
			conflicts = append(conflicts, k)
		}
	}
	return conflicts
}

func (a *App) GetLatestHash() (plumbing.Hash, error) {
	err := a.Fetch()
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		log.Fatal(err)
	}

	remote, err := a.repo.Remote("origin")
	if err != nil {
		log.Fatal(err)
	}

	refs, err := remote.List(&git.ListOptions{
		Auth: a.auth,
	})
	if err != nil {
		log.Fatal(err)
	}

	// find HEAD
	var target plumbing.ReferenceName
	m := make(map[plumbing.ReferenceName]plumbing.Hash)
	for _, ref := range refs {
		switch ref.Type() {
		case plumbing.SymbolicReference:
			if ref.Name() == "HEAD" {
				target = ref.Target()
			}
		case plumbing.HashReference:
			m[ref.Name()] = ref.Hash()
		}
	}

	if target == "" {
		return plumbing.Hash{}, errors.New("detached HEAD")
	}
	hash, ok := m[target]
	if !ok {
		return plumbing.Hash{}, git.ErrInvalidReference
	}
	return hash, nil
}

func (a *App) RemoteModifiedFiles() (map[string]struct{}, error) {
	// get local commit
	iter, err := a.repo.Log(&git.LogOptions{})
	if err != nil {
		return nil, err
	}

	c1, err := iter.Next()
	if err != nil {
		return nil, err
	}

	// get commit of origin/head
	hash, err := a.GetLatestHash()
	if err != nil {
		return nil, err
	}
	c2, err := a.repo.CommitObject(hash)
	if err != nil {
		return nil, err
	}

	// generate difference
	patch, err := c2.Patch(c1)
	if err != nil {
		return nil, err
	}

	files := make(map[string]struct{})
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		files[from.Path()] = struct{}{}
		files[to.Path()] = struct{}{}
	}
	return files, nil
}

func (a *App) Fetch() error {
	err := a.repo.Fetch(&git.FetchOptions{
		Auth: a.auth,
	})
	return err
}

func (a *App) ReadConfig() error {
	file, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&a.config)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) WriteConfig() error {
	data, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile("config.json", data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) OpenRepository() error {
	r, err := git.PlainOpen(a.config.WorkingDirectory)
	if err != nil {
		return err
	}
	a.repo = r
	// start background fetch routine
	a.startBackgroundFetch()
	a.startBackgroundStatus()
	a.startBackgroundConflictDetection()
	return nil
}

func (a *App) CloneRepository(repoURL string) {
	before, _, found := strings.Cut(repoURL, ".git")
	if !found {
		log.Fatal(fmt.Errorf("invalid repo URL"))
	}
	s := strings.Split(before, "/")
	name := strings.TrimSpace(s[len(s)-1])
	//dir := filepath.Join(a.WorkingDirectory, name)
	tempDir, err := os.MkdirTemp(a.config.WorkingDirectory, name)
	if err != nil {
		log.Fatal(err)
	}
	//if err := os.Mkdir(dir, os.ModePerm); err != nil {
	//	log.Fatal(err)
	//}
	a.config.WorkingDirectory = tempDir
	repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      repoURL,
		Auth:     a.auth,
		Progress: os.Stdout,
	})
	if err != nil {
		log.Fatal(err)
	}
	a.repo = repo
	// start background fetch routine
	a.startBackgroundFetch()
	a.startBackgroundStatus()
	a.startBackgroundConflictDetection()
	a.WriteConfig()
}

func (a *App) OpenFile() string {
	defaultDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		ShowHiddenFiles:  true,
		DefaultDirectory: defaultDir,
	})
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func (a *App) UseHTTPS(patFile string) {
	content := a.ReadFile(patFile)
	content = strings.TrimSpace(content)
	a.auth = &gitHTTP.BasicAuth{
		Username: "username",
		Password: content,
	}
	a.config.PersonalAccessToken = content
	a.WriteConfig()
}

func (a *App) ReadFile(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func (a *App) Add(files []string) {
	w, err := a.repo.Worktree()
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		_, err = w.Add(f)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (a *App) Commit(msg string) {
	w, err := a.repo.Worktree()
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Commit(msg, &git.CommitOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) Push() {
	err := a.repo.Push(&git.PushOptions{
		Auth: a.auth,
	})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return
	}
	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) Status() (*git.Status, error) {
	w, err := a.repo.Worktree()
	if err != nil {
		return nil, err
	}
	status, err := w.Status()
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (a *App) Save(username string, email string) error {
	a.config.Username = username
	a.config.Email = email
	a.WriteConfig()
	return nil
}

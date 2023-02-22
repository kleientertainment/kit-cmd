package main

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"log"
	"os"
	"strings"
)

func (a *App) PullWithAbort() {
	err := a.Pull()
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			fmt.Printf("%s\n", err)
			return
		}
		fmt.Printf("%s\n", err)
	}
}

func (a *App) OpenRepository(dir string) error {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}
	a.repo = r
	return nil
}

func (a *App) Pull() error {
	w, err := a.repo.Worktree()
	if err != nil {
		log.Fatal(err)
	}
	err = w.Pull(&git.PullOptions{
		Auth: a.auth,
	})
	if err != nil {
		//if errors.Is(err, git.NoErrAlreadyUpToDate) {
		//	return err
		//}
		return err
	}
	return nil
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

func (a *App) CloneRepository(repoURL string) {
	before, _, found := strings.Cut(repoURL, ".git")
	if !found {
		log.Fatal(fmt.Errorf("invalid repo URL"))
	}
	s := strings.Split(before, "/")
	name := strings.TrimSpace(s[len(s)-1])
	//dir := filepath.Join(a.RepoDirectory, name)
	tempDir, err := os.MkdirTemp(a.config.RepoDirectory, name)
	if err != nil {
		log.Fatal(err)
	}
	//if err := os.Mkdir(dir, os.ModePerm); err != nil {
	//	log.Fatal(err)
	//}
	a.config.RepoDirectory = tempDir
	repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      repoURL,
		Auth:     a.auth,
		Progress: os.Stdout,
	})
	if err != nil {
		log.Fatal(err)
	}
	a.repo = repo
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

func (a *App) Push() error {
	err := a.repo.Push(&git.PushOptions{
		Auth: a.auth,
	})
	if err != nil {
		return err
	}
	return nil
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

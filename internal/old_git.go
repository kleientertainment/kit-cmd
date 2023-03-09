package internal

/*
import (
	"fmt"
	"github.com/go-internal/go-internal/v5"
	"github.com/go-internal/go-internal/v5/plumbing"
	"log"
	"os"
	"os/exec"
	"strings"
)

func (a *App) StatusCmd() (string, error) {
	cmd := exec.Command("cmd", "/C", "internal", "status")
	output, err := cmdWrapper(cmd, app.config.RepoDirectory)
	if err != nil {
		return "", err
	}
	return output, nil
}

func PullCmd(param string) (string, error) {
	var err error
	var cmd *exec.Cmd

	cmd = exec.Command("cmd", "/C", "internal", "pull", param)
	output, err := cmdWrapper(cmd, app.config.RepoDirectory)
	if err != nil {
		return "", err
	}
	return output, nil
}

func AbortMerge() error {
	cmd := exec.Command("internal", "merge", "--abort")
	_, err := cmdWrapper(cmd, app.config.RepoDirectory)
	if err != nil {
		return err
	}
	return nil
}

func cmdWrapper(cmd *exec.Cmd, dir string) (string, error) {
	cmd.Dir = dir
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

func (a *App) PullWithAbort() error {
	err := a.Pull()
	if err != nil {
		fmt.Printf("%s\n", err)
		if errors.Is(err, internal.NoErrAlreadyUpToDate) {
			return nil
		}
		return err
	}
	return nil
}

func (a *App) OpenRepository(dir string) error {
	r, err := internal.PlainOpen(dir)
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
	err = w.Pull(&internal.PullOptions{
		//Auth: a.auth,
	})
	if err != nil {
		//if errors.Is(err, internal.NoErrAlreadyUpToDate) {
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
	if err != nil && !errors.Is(err, internal.NoErrAlreadyUpToDate) {
		log.Fatal(err)
	}

	remote, err := a.repo.Remote("origin")
	if err != nil {
		log.Fatal(err)
	}

	refs, err := remote.List(&internal.ListOptions{
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
		return plumbing.Hash{}, internal.ErrInvalidReference
	}
	return hash, nil
}

func (a *App) RemoteModifiedFiles() (map[string]struct{}, error) {
	// get local commit
	iter, err := a.repo.Log(&internal.LogOptions{})
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
	err := a.repo.Fetch(&internal.FetchOptions{
		Auth: a.auth,
	})
	return err
}

func (a *App) CloneRepository(repoURL string) {
	before, _, found := strings.Cut(repoURL, ".internal")
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
	repo, err := internal.PlainClone(tempDir, false, &internal.CloneOptions{
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
	_, err = w.Commit(msg, &internal.CommitOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) Push() error {
	err := a.repo.Push(&internal.PushOptions{
		Auth: a.auth,
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *App) Status() (*internal.Status, error) {
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
*/

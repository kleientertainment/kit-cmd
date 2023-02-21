package main

import (
	gitHTTP "github.com/go-git/go-git/v5/plumbing/transport/http"
	"strings"
)

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

func (a *App) BasicAuth() error {
	a.auth = &gitHTTP.BasicAuth{
		Username: a.config.Username,
		Password: a.config.PersonalAccessToken,
	}
	return nil
}

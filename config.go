package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/term"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

var config struct {
	repoDirectory       string
	personalAccessToken string
	username            string
	email               string
	serverDomain        string
}

type Config struct {
	RepoDirectory       string
	PersonalAccessToken string
	Username            string
	Email               string
	ServerDomain        string
}

func initFlags() {
	flag.StringVar(&config.repoDirectory, "repoDirectory", "", "repository directory root")
	flag.StringVar(&config.personalAccessToken, "pat", "", "personal access token")
	flag.StringVar(&config.username, "username", "", "username")
	flag.StringVar(&config.email, "email", "", "author email")
	flag.StringVar(&config.serverDomain, "serverDomain", "https://git.klei.com", "server domain")
	flag.Parse()
}

func newConfig() *Config {
	return &Config{
		RepoDirectory:       config.repoDirectory,
		PersonalAccessToken: config.personalAccessToken,
		Username:            config.username,
		Email:               config.email,
		ServerDomain:        config.serverDomain,
	}
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

func (a *App) Save(username string, email string) error {
	a.config.Username = username
	a.config.Email = email
	a.WriteConfig()
	return nil
}

func ExecutableDirectory() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func RepoDirectory() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dirname = filepath.Join(dirname, "Work", "go-git-test")
	return dirname
}

func credentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Gitea Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("Enter Gitea Personal Access Token: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}
	fmt.Println() // get rid of floating lack of newline
	password := string(bytePassword)
	return username, password, nil
}

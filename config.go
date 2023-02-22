package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golang.org/x/term"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type Config struct {
	RepoDirectory       string
	PersonalAccessToken string
	Username            string
	Email               string
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
	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

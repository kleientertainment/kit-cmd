package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	WorkingDirectory    string
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

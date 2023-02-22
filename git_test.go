package main

import (
	"log"
	"testing"
)

func setupSuite(t testing.T) func(t testing.T) {
	log.Println("setup test suite")
	initializeApplication()
	// Return a function to teardown the test
	return func(t testing.T) {
		log.Println("teardown test suite")
	}
}

func TestPull(t *testing.T) {
	initializeApplication()
	app.PullWithAbort()
}

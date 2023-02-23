package main

import (
	"testing"
)

func TestPull(t *testing.T) {
	initializeApplication()

	err := app.PullWithAbort()
	if err != nil {
		t.Fatal(err)
	}
}

/*func setupSuite(t testing.T) func(t testing.T) {
	log.Println("setup test suite")
	initializeApplication()

	// Return a function to teardown the test
	return func(t testing.T) {
		log.Println("teardown test suite")
	}
}*/

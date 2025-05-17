package tests

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Initialize the test database
	log.Println("Setting up test database...")
	TestDB = InitTestDB()
	
	// Run all tests
	log.Println("Running tests...")
	exitCode := m.Run()
	
	// Clean up any test resources if needed
	log.Println("Test execution finished")
	
	os.Exit(exitCode)
}

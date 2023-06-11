package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestExtractAccountsFromFile(t *testing.T) {
	// Create a temporary test file
	content := `
2015/10/12 Exxon
    Expenses:Auto:Gas         $10.00
    Liabilities:MasterCard   $-10.00
`
	tmpfile, err := ioutil.TempFile("", "testledger.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary test file: %v", err)
	}
	defer tmpfile.Close()
	defer removeFile(tmpfile.Name())

	// Write test content to the temporary file
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temporary test file: %v", err)
	}

	// Call the function being tested
	accounts, err := extractAccountsFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Extracting accounts failed: %v", err)
	}

	// Expected accounts
	expectedAccounts := []string{
		"Expenses:Auto:Gas",
		"Liabilities:MasterCard",
	}

	// Check if the extracted accounts match the expected accounts
	if len(accounts) != len(expectedAccounts) {
		t.Errorf("Number of extracted accounts does not match the expected count")
	}

	for i, account := range accounts {
		if account != expectedAccounts[i] {
			t.Errorf("Extracted account does not match the expected account at index %d", i)
		}
	}
}

// Helper function to remove a file
func removeFile(filePath string) {
	if err := os.Remove(filePath); err != nil {
		log.Printf("Failed to remove file %s: %v", filePath, err)
	}
}

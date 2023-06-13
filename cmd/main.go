package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ledger-wizard <file>",
		Short: "A wizard for adding a new ledger entry",
		RunE:  runWizard,
		Args:  cobra.ExactArgs(1),
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runWizard(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	reader := bufio.NewReader(os.Stdin)

	date := promptForDate(reader)
	description := promptForDescription(reader)
	toAccount := promptForAccount(filePath)
	fromAccount := promptForAccount(filePath)
	amount := promptForAmount()

	ledgerEntry := buildLedgerEntry(date, description, toAccount, fromAccount, amount)
	ledgerFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer ledgerFile.Close()

	_, err = ledgerFile.WriteString(ledgerEntry + "\n")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Ledger entry added successfully.")

	return nil
}

func promptForAmount() string {
	prompt := promptui.Prompt{
		Label:   "Amount",
		Default: "€10",
	}

	result, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}

	return result
}

func promptForDate(reader *bufio.Reader) string {
	prompt := promptui.Prompt{
		Label:   "Date (YYYY/MM/DD)",
		Default: time.Now().Format("2006/01/02"),
	}

	result, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}

	return result
}

func promptForDescription(reader *bufio.Reader) string {
	prompt := promptui.Prompt{
		Label: "Description",
	}

	result, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}

	return result
}

func extractAccountsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	accounts := make(map[string]struct{})

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and lines starting with '#'
		}

		match := regexp.MustCompile(`^\s*([\w:]+)\s{2,}`).FindStringSubmatch(line)
		if len(match) == 2 {
			account := strings.TrimSpace(match[1])

			// Add account to the map
			accounts[account] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Extract unique accounts from the map
	uniqueAccounts := make([]string, 0, len(accounts))
	for account := range accounts {
		uniqueAccounts = append(uniqueAccounts, account)
	}

	return uniqueAccounts, nil
}

func promptForAccount(filepath string) string {
	accountLines, err := extractAccountsFromFile(filepath)
	if err != nil {
		panic(err)
	}

	fzfInput := strings.Join(accountLines, "\n")

	cmd := exec.Command("fzf", "--multi", "--print-query")
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Stdin = strings.NewReader(fzfInput)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(stdout)

	var accounts []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		account := strings.TrimSpace(line)
		if account != "" {
			accounts = append(accounts, account)
		}
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	return accounts[0]
}

func buildLedgerEntry(date, description string, toAccount string, fromAccount string, amount string) string {
	entryBuilder := strings.Builder{}

	entryBuilder.WriteString(date)
	entryBuilder.WriteString(" ")
	entryBuilder.WriteString(description)
	entryBuilder.WriteString("\n")

	entryBuilder.WriteString("\t")
	entryBuilder.WriteString(toAccount)
	entryBuilder.WriteString(fmt.Sprintf("\t\t%v\n", amount))

	entryBuilder.WriteString("\t")
	entryBuilder.WriteString(fromAccount)
	entryBuilder.WriteString(fmt.Sprintf("\t\t-%v\n", amount))

	return entryBuilder.String()
}

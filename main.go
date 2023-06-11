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
		Use:   "ledger-wizard",
		Short: "A wizard for adding a new ledger entry",
		RunE:  runWizard,
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runWizard(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	date := promptForDate(reader)
	description := promptForDescription(reader)
	toAccount := promptForAccount()
	fromAccount := promptForAccount()
	amount := promptForAmount()

	ledgerEntry := buildLedgerEntry(date, description, toAccount, fromAccount, amount)

	ledgerFile, err := os.OpenFile("ledger.txt", os.O_WRONLY|os.O_APPEND, 0644)
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

	var accounts []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		match := regexp.MustCompile(`^\s*([\w:]+)\s{2,}`).FindStringSubmatch(line)
		if len(match) == 2 {
			account := strings.TrimSpace(match[1])
			found := false
			for _, searchedAcc := range accounts {
				if searchedAcc == account {
					found = true
				}
			}
			if !found {
				accounts = append(accounts, account)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func promptForAccount() string {
	accountLines, err := extractAccountsFromFile("ledger.txt")
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
	cmd.Start()

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

	cmd.Wait()

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

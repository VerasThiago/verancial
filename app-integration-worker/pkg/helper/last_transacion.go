package helper

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/models"
)

// TODO: Create a interface and implement it to every app method
func GetLastTransaction(financialAppCredentials *models.FinancialAppCredentials, bankId constants.BankId) (time.Time, error) {
	var err error
	var stdout, stderr bytes.Buffer
	var lastTransactionTime time.Time

	pythonPath := "/Users/veras/go/src/verancial/python-scripts/env/bin/python3"
	filePath := "/Users/veras/go/src/verancial/python-scripts/get_last_transaction_budgetbakers.py"

	cmd := exec.Command(pythonPath, filePath, financialAppCredentials.Login, financialAppCredentials.Password, financialAppCredentials.Metadata[bankId])
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		fmt.Println("Script error:", stderr.String())
		return time.Time{}, err
	}

	layout := "January 2 2006"
	timeStr := stdout.String()[:len(stdout.String())-1] // Remove escaped hexadecimal Line Feed
	fullDateStr := fmt.Sprintf("%s %d", timeStr, time.Now().Year())

	if lastTransactionTime, err = time.Parse(layout, fullDateStr); err != nil {
		fmt.Println("error parsing date:", err)
		return time.Time{}, err
	}

	return lastTransactionTime, nil
}

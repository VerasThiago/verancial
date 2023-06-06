package budgetbakers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/verasthiago/verancial/app-integration-worker/pkg/generators/helper"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/types"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/models"
)

type BudgetBakers struct{}

func (b BudgetBakers) Generate(transactions []*models.Transaction) (types.AppReport, error) {
	var csv types.AppReport = types.AppReport{}
	for _, t := range transactions {
		bb := models.BudgetBakers{
			Date:        t.Date.In(time.UTC),
			Amount:      t.Amount,
			Description: fmt.Sprintf("%v | %v", t.Category, t.Description),
			Payee:       t.Payee,
			Currency:    t.Currency,
		}
		row := []string{bb.Date.Format("2006-01-02"), fmt.Sprintf("%f", bb.Amount), bb.Description, bb.Payee, bb.Currency}
		csv = append(csv, row)
	}

	return csv, nil
}

func (b BudgetBakers) Submit(user *models.User, appReport types.AppReport) error {
	fileName := fmt.Sprintf("%v.csv", user.Email)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range appReport {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (b BudgetBakers) GetLastTransaction(financialAppCredentials *models.FinancialAppCredentials, bankId constants.BankId) (time.Time, error) {
	var err error
	var stdout, stderr bytes.Buffer
	var lastTransactionTime time.Time
	var pythonScriptsPath string

	if pythonScriptsPath, err = helper.GetPythonScriptsPath(); err != nil {
		return time.Time{}, err
	}
	pythonPath := path.Join(pythonScriptsPath, "env/bin/python3")
	filePath := path.Join(pythonScriptsPath, "get_last_transaction_budgetbakers.py")

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

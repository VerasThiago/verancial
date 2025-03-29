package budgetbakers

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/verasthiago/verancial/app-integration-worker/pkg/generators/helper"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/types"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/models"
)

type BudgetBakers struct{}

func (b BudgetBakers) Generate(transactions []*models.Transaction) (types.AppReport, error) {
	var csv types.AppReport = types.AppReport{}
	var header []string = []string{"Date", "Amount", "Note", "Payee", "Currency"}
	csv = append(csv, header)
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
	fileName := helper.GetFileNameFromAppReport(appReport)
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

func (b BudgetBakers) GetLastTransaction(financialAppCredentials *models.FinancialAppCredentials, bankId constants.BankId, lastTransactionDate string) (time.Time, error) {
	var err error
	var lastTransactionTime time.Time

	layout := "January 2 2006"
	if lastTransactionTime, err = time.Parse(layout, lastTransactionDate); err != nil {
		return time.Time{}, err
	}

	return lastTransactionTime, nil
}

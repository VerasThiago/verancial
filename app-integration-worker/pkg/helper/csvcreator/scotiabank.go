package csvcreator

import (
	"fmt"
	"time"

	"github.com/verasthiago/verancial/shared/models"
)

func CreateCSVFromTransactions(transactions []*models.Transaction) ([][]string, error) {
	var csv [][]string = [][]string{}
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

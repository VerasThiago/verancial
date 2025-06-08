package firsttech

import (
	"encoding/csv"
	"io"
	"os"

	categoryguesser "github.com/verasthiago/verancial/data-process-worker/pkg/category-guesser"
	"github.com/verasthiago/verancial/data-process-worker/pkg/models/firsttech"
	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

type FirstTechReportProcessor struct{}

func (f FirstTechReportProcessor) LoadFromCSV(filePath string) ([]interface{}, error) {
	var err error
	var file *os.File
	var reader *csv.Reader
	var transactions []interface{}

	if file, err = os.Open(filePath); err != nil {
		return nil, err
	}

	defer file.Close()

	reader = csv.NewReader(file)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		transaction, err := f.ParseReportRecord(record)
		if err != nil {
			return nil, err
		}

		if transaction != nil {
			transactions = append(transactions, transaction)
		}
	}

	return transactions, nil
}

func (f FirstTechReportProcessor) Process(bankTransactions []interface{}, payload *types.ReportProcessQueuePayload, lastDbTransaction *models.Transaction) ([]*models.Transaction, error) {
	var transactions []*models.Transaction

	for _, bankTransaction := range bankTransactions {
		firstTechTransaction, ok := bankTransaction.(*firsttech.FirstTech)
		if !ok {
			return nil, errors.GenericError{
				Code:    errors.STATUS_BAD_REQUEST,
				Type:    errors.GENERIC_ERROR.Type,
				Message: errors.GENERIC_ERROR.Message,
			}
		}

		// Use posting date for comparison and transaction creation
		if firstTechTransaction.PostingDate.After(lastDbTransaction.Date) {
			payee := firstTechTransaction.Description
			category, err := categoryguesser.GuessCategory(payee)
			if err != nil {
				return nil, err
			}

			transactions = append(transactions, &models.Transaction{
				UserId:      payload.UserId,
				Date:        firstTechTransaction.PostingDate,
				Amount:      firstTechTransaction.Amount,
				Payee:       payee,
				Description: firstTechTransaction.Description,
				Category:    category,
				Currency:    "USD",
				BankId:      payload.BankId,
				Metadata:    nil,
			})
		}
	}

	return transactions, nil
}

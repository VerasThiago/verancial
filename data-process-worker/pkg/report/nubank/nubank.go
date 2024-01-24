package nubank

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/verasthiago/verancial/data-process-worker/pkg/models/nubank"
	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

type NubankReportProcessor struct{}

func (n NubankReportProcessor) LoadFromCSV(filePath string) ([]interface{}, error) {
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

		transaction, err := n.ParseReportRecord(record)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (s NubankReportProcessor) Process(bankTransactions []interface{}, payload *types.ReportProcessQueuePayload, lastDbTransaction *models.Transaction) ([]*models.Transaction, error) {
	var transactions []*models.Transaction

	for _, bankTransaction := range bankTransactions {
		nubankTransaction, ok := bankTransaction.(*nubank.Nubank)
		if !ok {
			return nil, errors.GenericError{
				Code:    errors.STATUS_BAD_REQUEST,
				Type:    errors.GENERIC_ERROR.Type,
				Message: errors.GENERIC_ERROR.Message,
			}
		}

		if nubankTransaction.Date.After(lastDbTransaction.Date) {
			transactions = append(transactions, &models.Transaction{
				UserId:      payload.UserId,
				Date:        (*nubankTransaction).Date,
				Amount:      (*nubankTransaction).Amount,
				Payee:       (*nubankTransaction).Payee,
				Description: (*nubankTransaction).Description,
				//TODO: Use AI to guess current category
				Category: "",
				// TODO: Get currency from user info (?)
				Currency: "BRL",
				BankId:   payload.BankId,
				Metadata: nil,
			})
		}
	}

	return transactions, nil
}

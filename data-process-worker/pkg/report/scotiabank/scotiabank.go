package scotiabank

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/verasthiago/verancial/data-process-worker/pkg/models/scotiabank"
	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

type ScotiaBankReportProcessor struct{}

func (s ScotiaBankReportProcessor) LoadFromCSV(filePath string) ([]interface{}, error) {
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

		scotiaBankTransaction, err := s.ParseReportRecord(record)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, scotiaBankTransaction)
	}

	return transactions, nil
}

func (s ScotiaBankReportProcessor) Process(bankTransactions []interface{}, payload *types.ReportProcessQueuePayload, lastDbTransaction *models.Transaction) ([]*models.Transaction, error) {
	var transactions []*models.Transaction

	for _, bankTransaction := range bankTransactions {
		sbTransacion, ok := bankTransaction.(*scotiabank.ScotiaBank)
		if !ok {
			return nil, errors.GenericError{
				Code:    errors.STATUS_BAD_REQUEST,
				Type:    errors.GENERIC_ERROR.Type,
				Message: errors.GENERIC_ERROR.Message,
			}
		}

		if sbTransacion.Date.After(lastDbTransaction.Date) {
			transactions = append(transactions, &models.Transaction{
				UserId:      payload.UserId,
				Date:        sbTransacion.Date,
				Amount:      sbTransacion.Amount,
				Payee:       sbTransacion.Payee,
				Description: sbTransacion.Description,
				//TODO: Use AI to guess current category
				Category: "",
				// TODO: Get currency from user info (?)
				Currency: "CAD",
				BankId:   payload.BankId,
				Metadata: nil,
			})
		}
	}

	return transactions, nil
}

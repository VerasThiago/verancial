package nubank

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/worker/pkg/models/nubank"
)

type NubankReportProcessor struct{}

func (n NubankReportProcessor) LoadFromCSV(filePath string) (error, []interface{}) {
	var err error
	var file *os.File
	var reader *csv.Reader
	var transactions []interface{}

	if file, err = os.Open(filePath); err != nil {
		return err, nil
	}
	defer file.Close()

	reader = csv.NewReader(file)

	// Skip header
	if _, err = reader.Read(); err != nil {
		return err, nil
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err, nil
		}

		err, transaction := n.ParseReportRecord(record)
		if err != nil {
			return err, nil
		}

		transactions = append(transactions, transaction)
	}

	return nil, transactions
}

func (s NubankReportProcessor) Process(bankTransactions []interface{}) (error, []*models.Transaction) {
	var transactions []*models.Transaction

	for _, bankTransaction := range bankTransactions {
		nubankTransaction, ok := bankTransaction.(*nubank.Nubank)
		if !ok {
			return errors.GenericError{
				Code:    errors.STATUS_BAD_REQUEST,
				Type:    errors.GENERIC_ERROR.Type,
				Message: errors.GENERIC_ERROR.Message,
			}, nil
		}

		transactions = append(transactions, &models.Transaction{
			Date:        (*nubankTransaction).Date,
			Amount:      (*nubankTransaction).Amount,
			Payee:       (*nubankTransaction).Payee,
			Description: (*nubankTransaction).Description,
			//TODO: Use AI to guess current category
			Category: "",
			Metadata: make(map[string]string),
		})
	}

	return nil, transactions
}

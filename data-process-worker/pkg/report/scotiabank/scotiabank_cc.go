package scotiabank

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/worker/pkg/models/scotiabank"
)

type ScotiaBankCCReportProcessor struct{}

func (s ScotiaBankCCReportProcessor) LoadFromCSV(filePath string) (error, []interface{}) {
	var err error
	var file *os.File
	var reader *csv.Reader
	var transactions []interface{}

	if file, err = os.Open(filePath); err != nil {
		return err, nil
	}
	defer file.Close()

	reader = csv.NewReader(file)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err, nil
		}

		err, transaction := s.ParseReportRecord(record)
		if err != nil {
			return err, nil
		}

		transactions = append(transactions, transaction)
	}

	return nil, transactions
}

func (s ScotiaBankCCReportProcessor) Process(bankTransactions []interface{}) (error, []*models.Transaction) {
	var transactions []*models.Transaction

	for _, bankTransaction := range bankTransactions {
		sbTransacion, ok := bankTransaction.(*scotiabank.ScotiaBank)
		if !ok {
			return errors.GenericError{
				Code:    errors.STATUS_BAD_REQUEST,
				Type:    errors.GENERIC_ERROR.Type,
				Message: errors.GENERIC_ERROR.Message,
			}, nil
		}

		transactions = append(transactions, &models.Transaction{
			Date:        sbTransacion.Date,
			Amount:      sbTransacion.Amount,
			Payee:       sbTransacion.Payee,
			Description: sbTransacion.Description,
			//TODO: Use AI to guess current category
			Category: "",
			Metadata: make(map[string]string),
		})
	}

	return nil, transactions
}

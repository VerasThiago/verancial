package wise

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/worker/pkg/models/wise"
)

type WiseReportProcessor struct{}

func (s WiseReportProcessor) LoadFromCSV(filePath string) (error, []interface{}) {
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

		err, transaction := s.ParseReportRecord(record)
		if err != nil {
			return err, nil
		}

		transactions = append(transactions, transaction)
	}

	return nil, transactions
}

func (s WiseReportProcessor) Process(bankTransactions []interface{}) (error, []*models.Transaction) {
	var transactions []*models.Transaction

	for _, bankTransaction := range bankTransactions {
		wiseTransaction, ok := bankTransaction.(*wise.Wise)
		if !ok {
			return errors.GenericError{
				Code:    errors.STATUS_BAD_REQUEST,
				Type:    errors.GENERIC_ERROR.Type,
				Message: errors.GENERIC_ERROR.Message,
			}, nil
		}

		transactions = append(transactions, &models.Transaction{
			Date:        wiseTransaction.FinishedOn,
			Amount:      wiseTransaction.TargetAmountAfterFees,
			Payee:       wiseTransaction.TargetName,
			Description: wiseTransaction.Reference,
			//TODO: Use AI to guess current category
			Category: "",
			Metadata: map[string]string{
				"TransactionID":         wiseTransaction.TransactionID,
				"Status":                wiseTransaction.Status,
				"Direction":             wiseTransaction.Direction,
				"CreatedOn":             wiseTransaction.CreatedOn.GoString(),
				"SourceFeeAmount":       ParseFloatToString(wiseTransaction.SourceFeeAmount),
				"SourceFeeCurrency":     wiseTransaction.SourceFeeCurrency,
				"TargetFeeAmount":       ParseFloatToString(wiseTransaction.TargetFeeAmount),
				"TargetFeeCurrency":     wiseTransaction.TargetFeeCurrency,
				"SourceName":            wiseTransaction.SourceName,
				"SourceAmountAfterFees": ParseFloatToString(wiseTransaction.SourceAmountAfterFees),
				"SourceCurrency":        wiseTransaction.SourceCurrency,
				"TargetCurrency":        wiseTransaction.TargetCurrency,
				"ExchangeRate":          ParseFloatToString(wiseTransaction.ExchangeRate),
				"Batch":                 wiseTransaction.Batch,
			},
		})
	}

	return nil, transactions
}

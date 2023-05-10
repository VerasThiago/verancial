package nubank

import (
	"strings"
	"time"

	"github.com/verasthiago/verancial/data-process-worker/pkg/helper"
	"github.com/verasthiago/verancial/data-process-worker/pkg/models/nubank"
)

func getPayee(record string) string {
	splited := strings.Split(record, " - ")
	return splited[1]
}

func (s NubankReportProcessor) ParseReportRecord(record []string) (*nubank.Nubank, error) {
	var err error
	var date time.Time
	var amount float32

	date, err = time.Parse("02/01/2006", record[0])
	if err != nil {
		return nil, err
	}
	amount, err = helper.ParseFloat(record[1])
	if err != nil {
		return nil, err
	}

	return &nubank.Nubank{
		Date:        date,
		Amount:      float32(amount),
		Description: record[3],
		Payee:       getPayee(record[3]),
	}, nil
}

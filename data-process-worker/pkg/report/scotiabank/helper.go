package scotiabank

import (
	"regexp"
	"strings"
	"time"

	"github.com/verasthiago/verancial/data-process-worker/pkg/helper"
	"github.com/verasthiago/verancial/data-process-worker/pkg/models/scotiabank"
)

// TODO: Investigate why record[2] is always empty (after downloading)
func (s ScotiaBankReportProcessor) ParseReportRecord(record []string) (*scotiabank.ScotiaBank, error) {
	var err error
	var date time.Time
	var amount float32
	var spacesRegex = regexp.MustCompile(`\s+`)

	if date, err = time.Parse("1/2/2006", record[0]); err != nil {
		return nil, err
	}

	if amount, err = helper.ParseFloat(record[1]); err != nil {
		return nil, err
	}

	return &scotiabank.ScotiaBank{
		Date:        date,
		Amount:      float32(amount),
		Description: strings.TrimSpace(spacesRegex.ReplaceAllString(record[3], " ")),
		Payee:       strings.TrimSpace(spacesRegex.ReplaceAllString(record[4], " ")),
	}, nil
}

func (s ScotiaBankCCReportProcessor) ParseReportRecord(record []string) (*scotiabank.ScotiaBank, error) {
	var err error
	var date time.Time
	var amount float32
	var spacesRegex = regexp.MustCompile(`\s+`)

	date, err = time.Parse("1/2/2006", record[0])
	if err != nil {
		return nil, err
	}
	amount, err = helper.ParseFloat(record[2])
	if err != nil {
		return nil, err
	}

	return &scotiabank.ScotiaBank{
		Date:        date,
		Payee:       strings.TrimSpace(spacesRegex.ReplaceAllString(record[1], " ")),
		Amount:      float32(amount),
		Description: "",
	}, nil
}

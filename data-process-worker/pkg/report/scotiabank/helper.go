package scotiabank

import (
	"regexp"
	"strings"
	"time"

	"github.com/verasthiago/verancial/data-process-worker/pkg/helper"
	"github.com/verasthiago/verancial/data-process-worker/pkg/models/scotiabank"
)

// spacesRegex is compiled once at package init instead of on every call to
// ParseReportRecord, which previously recompiled it for every row of every
// imported CSV statement.
var spacesRegex = regexp.MustCompile(`\s+`)

func (s ScotiaBankReportProcessor) ParseReportRecord(record []string) (*scotiabank.ScotiaBank, error) {
	var err error
	var date time.Time
	var amount float32

	if date, err = time.Parse("2006-01-02", record[1]); err != nil {
		return nil, err
	}

	if amount, err = helper.ParseAmountFloat(record[5]); err != nil {
		return nil, err
	}

	return &scotiabank.ScotiaBank{
		Date:        date,
		Amount:      float32(amount),
		Description: strings.TrimSpace(spacesRegex.ReplaceAllString(record[2], " ")),
		Payee:       strings.TrimSpace(spacesRegex.ReplaceAllString(record[3], " ")),
	}, nil
}

func (s ScotiaBankCCReportProcessor) ParseReportRecord(record []string) (*scotiabank.ScotiaBank, error) {
	var err error
	var date time.Time
	var amount float32

	if date, err = time.Parse("2006-01-02", record[1]); err != nil {
		return nil, err
	}

	if amount, err = helper.ParseAmountFloat(record[6]); err != nil {
		return nil, err
	}

	return &scotiabank.ScotiaBank{
		Date:        date,
		Amount:      float32(amount),
		Description: strings.TrimSpace(spacesRegex.ReplaceAllString(record[3], " ")),
		Payee:       strings.TrimSpace(spacesRegex.ReplaceAllString(record[2], " ")),
	}, nil
}

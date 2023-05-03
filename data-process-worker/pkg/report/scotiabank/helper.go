package scotiabank

import (
	"regexp"
	"strings"
	"time"

	"github.com/verasthiago/verancial/worker/pkg/helper"
	"github.com/verasthiago/verancial/worker/pkg/models/scotiabank"
)

// TODO: Investigate why record[2] is always empty (after downloading)
func (s ScotiaBankReportProcessor) ParseReportRecord(record []string) (error, *scotiabank.ScotiaBank) {
	var err error
	var date time.Time
	var amount float32
	var spacesRegex = regexp.MustCompile(`\s+`)

	if date, err = time.Parse("1/2/2006", record[0]); err != nil {
		return err, nil
	}

	if amount, err = helper.ParseFloat(record[1]); err != nil {
		return err, nil
	}

	return nil, &scotiabank.ScotiaBank{
		Date:        date,
		Amount:      float32(amount),
		Description: strings.TrimSpace(spacesRegex.ReplaceAllString(record[3], " ")),
		Payee:       strings.TrimSpace(spacesRegex.ReplaceAllString(record[4], " ")),
	}
}

func (s ScotiaBankCCReportProcessor) ParseReportRecord(record []string) (error, *scotiabank.ScotiaBank) {
	var err error
	var date time.Time
	var amount float32
	var spacesRegex = regexp.MustCompile(`\s+`)

	date, err = time.Parse("1/2/2006", record[0])
	if err != nil {
		return err, nil
	}
	amount, err = helper.ParseFloat(record[2])
	if err != nil {
		return err, nil
	}

	return nil, &scotiabank.ScotiaBank{
		Date:        date,
		Payee:       strings.TrimSpace(spacesRegex.ReplaceAllString(record[1], " ")),
		Amount:      float32(amount),
		Description: "",
	}
}

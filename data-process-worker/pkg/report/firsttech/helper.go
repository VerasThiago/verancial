package firsttech

import (
	"time"

	"github.com/verasthiago/verancial/data-process-worker/pkg/helper"
	"github.com/verasthiago/verancial/data-process-worker/pkg/models/firsttech"
)

func (f FirstTechReportProcessor) ParseReportRecord(record []string) (*firsttech.FirstTech, error) {
	var err error
	var postingDate, effectiveDate time.Time
	var amount, balance float32

	if record[0] == "Transaction ID" {
		return nil, nil
	}

	if postingDate, err = time.Parse("1/2/2006", record[1]); err != nil {
		return nil, err
	}

	if effectiveDate, err = time.Parse("1/2/2006", record[2]); err != nil {
		return nil, err
	}

	if amount, err = helper.ParseAmountFloat(record[4]); err != nil {
		return nil, err
	}

	if balance, err = helper.ParseAmountFloat(record[10]); err != nil {
		return nil, err
	}

	return &firsttech.FirstTech{
		TransactionID:       record[0],
		PostingDate:         postingDate,
		EffectiveDate:       effectiveDate,
		TransactionType:     record[3],
		Amount:              amount,
		CheckNumber:         record[5],
		ReferenceNumber:     record[6],
		Description:         record[7],
		TransactionCategory: record[8],
		Type:                record[9],
		Balance:             balance,
		Memo:                record[11],
		ExtendedDescription: record[12],
	}, nil
}

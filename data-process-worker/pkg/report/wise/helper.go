package wise

import (
	"io"
	"strconv"
	"time"

	"github.com/verasthiago/verancial/data-process-worker/pkg/helper"
	"github.com/verasthiago/verancial/data-process-worker/pkg/models/wise"
)

func (s WiseReportProcessor) ParseReportRecord(record []string) (*wise.Wise, error) {
	createdOn, err := time.Parse("2006-01-02 15:04:05", record[3])
	if err != nil {
		return nil, err
	}

	finishedOn, err := time.Parse("2006-01-02 15:04:05", record[4])
	if err != nil {
		return nil, err
	}

	sourceFeeAmount, err := helper.ParseFloat(record[5])
	if err != nil && err != io.EOF {
		return nil, err
	}

	targetFeeAmount, err := helper.ParseFloat(record[7])
	if err != nil && err != io.EOF {
		return nil, err
	}

	sourceAmountAfterFees, err := helper.ParseFloat(record[10])
	if err != nil && err != io.EOF {
		return nil, err
	}

	targetAmountAfterFees, err := helper.ParseFloat(record[13])
	if err != nil && err != io.EOF {
		return nil, err
	}

	exchangeRate, err := helper.ParseFloat(record[15])
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &wise.Wise{
		TransactionID:         record[0],
		Status:                record[1],
		Direction:             record[2],
		CreatedOn:             createdOn,
		FinishedOn:            finishedOn,
		SourceFeeAmount:       sourceFeeAmount,
		SourceFeeCurrency:     record[6],
		TargetFeeAmount:       targetFeeAmount,
		TargetFeeCurrency:     record[8],
		SourceName:            record[9],
		SourceAmountAfterFees: sourceAmountAfterFees,
		SourceCurrency:        record[11],
		TargetName:            record[12],
		TargetAmountAfterFees: targetAmountAfterFees,
		TargetCurrency:        record[14],
		ExchangeRate:          exchangeRate,
		Reference:             record[16],
		Batch:                 record[17],
	}, nil

}

func ParseFloatToString(number float32) string {
	return strconv.FormatFloat(float64(number), 'f', -1, 32)
}

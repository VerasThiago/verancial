package report

import (
	"fmt"

	"github.com/verasthiago/verancial/shared/constants"
)

type BankReport interface {
	ProcessReport(filePath string) error
}

func GetReportProcessor(bankName constants.BankName) (error, BankReport) {
	switch bankName {
	case constants.ScotiaBank:
		return nil, ScotiaBankReportProcessor{}
	case constants.Wise:
		return nil, WiseReportProcessor{}
	default:
		return fmt.Errorf("Bank not supported"), nil
	}
}

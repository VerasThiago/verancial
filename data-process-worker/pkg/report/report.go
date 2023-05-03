package report

import (
	"fmt"

	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/worker/pkg/report/nubank"
	"github.com/verasthiago/verancial/worker/pkg/report/scotiabank"
	"github.com/verasthiago/verancial/worker/pkg/report/wise"
)

type BankReport interface {
	LoadFromCSV(csvPath string) (error, []interface{})
	Process(bankTransactions []interface{}) (error, []*models.Transaction)
}

func GetReportProcessor(bankName constants.BankID) (error, BankReport) {
	switch bankName {
	case constants.ScotiaBank:
		return nil, scotiabank.ScotiaBankReportProcessor{}
	case constants.ScotiaBankCC:
		return nil, scotiabank.ScotiaBankCCReportProcessor{}
	case constants.Nubank:
		return nil, nubank.NubankReportProcessor{}
	case constants.Wise:
		return nil, wise.WiseReportProcessor{}
	default:
		return fmt.Errorf("Bank not supported"), nil
	}
}

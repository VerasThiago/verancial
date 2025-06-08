package report

import (
	"fmt"

	"github.com/verasthiago/verancial/data-process-worker/pkg/report/firsttech"
	"github.com/verasthiago/verancial/data-process-worker/pkg/report/nubank"
	"github.com/verasthiago/verancial/data-process-worker/pkg/report/scotiabank"
	"github.com/verasthiago/verancial/data-process-worker/pkg/report/wise"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

type BankReport interface {
	LoadFromCSV(csvPath string) ([]interface{}, error)
	Process(bankTransactions []interface{}, payload *types.ReportProcessQueuePayload, lastDbTransaction *models.Transaction) ([]*models.Transaction, error)
}

func GetReportProcessor(bankName constants.BankId) (BankReport, error) {
	switch bankName {
	case constants.ScotiaBank:
		return scotiabank.ScotiaBankReportProcessor{}, nil
	case constants.ScotiaBankCC:
		return scotiabank.ScotiaBankCCReportProcessor{}, nil
	case constants.Nubank:
		return nubank.NubankReportProcessor{}, nil
	case constants.Wise:
		return wise.WiseReportProcessor{}, nil
	case constants.FirstTech:
		return firsttech.FirstTechReportProcessor{}, nil
	case constants.FirstTechCC:
		return firsttech.FirstTechCCReportProcessor{}, nil
	default:
		return nil, fmt.Errorf("bank not supported")
	}
}

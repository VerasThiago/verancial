package generators

import (
	"fmt"
	"time"

	"github.com/verasthiago/verancial/app-integration-worker/pkg/generators/budgetbakers"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/types"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/models"
)

type AppReport interface {
	Generate(transactions []*models.Transaction) (types.AppReport, error)
	Submit(user *models.User, appReport types.AppReport) error
	GetLastTransaction(financialAppCredentials *models.FinancialAppCredentials, bankId string, lastTransactionDate string) (time.Time, error)
}

func GetAppReportGenerator(appName constants.AppID) (AppReport, error) {
	switch appName {
	case constants.BudgetBakers:
		return budgetbakers.BudgetBakers{}, nil
	default:
		return nil, fmt.Errorf("app not supported")
	}
}

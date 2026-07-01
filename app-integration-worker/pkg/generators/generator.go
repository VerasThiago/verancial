package generators

import (
	"fmt"
	"time"

	"github.com/verasthiago/verancial/app-integration-worker/pkg/generators/budgetbakers"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/types"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/models"
)

//go:generate mockgen -source=generator.go -destination=mocks/mock_generator.go -package=mocks

type AppReport interface {
	Generate(transactions []*models.Transaction) (types.AppReport, error)
	Submit(user *models.User, appReport types.AppReport) error
	GetLastTransaction(financialAppCredentials *models.FinancialAppCredentials, bankId string, lastTransactionDate string) (time.Time, error)
}

// AppReportGeneratorFactory resolves the AppReport implementation for a
// given third-party app. It is injected via the builder so handlers can be
// tested with a fake factory instead of exercising the real generator.
type AppReportGeneratorFactory interface {
	GetAppReportGenerator(appName constants.AppID) (AppReport, error)
}

type DefaultAppReportGeneratorFactory struct{}

func (DefaultAppReportGeneratorFactory) GetAppReportGenerator(appName constants.AppID) (AppReport, error) {
	switch appName {
	case constants.BudgetBakers:
		return budgetbakers.BudgetBakers{}, nil
	default:
		return nil, fmt.Errorf("app not supported")
	}
}

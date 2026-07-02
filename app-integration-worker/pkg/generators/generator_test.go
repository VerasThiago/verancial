package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/generators/budgetbakers"
	"github.com/verasthiago/verancial/shared/constants"
)

func TestDefaultAppReportGeneratorFactory_GetAppReportGenerator_BudgetBakers(t *testing.T) {
	factory := DefaultAppReportGeneratorFactory{}

	generator, err := factory.GetAppReportGenerator(constants.BudgetBakers)

	require.NoError(t, err)
	require.NotNil(t, generator)
	assert.IsType(t, budgetbakers.BudgetBakers{}, generator)
}

func TestDefaultAppReportGeneratorFactory_GetAppReportGenerator_Unsupported(t *testing.T) {
	factory := DefaultAppReportGeneratorFactory{}

	generator, err := factory.GetAppReportGenerator(constants.AppID("unsupported-app-id"))

	assert.Error(t, err)
	assert.Nil(t, generator)
	assert.EqualError(t, err, "app not supported")
}

func TestDefaultAppReportGeneratorFactory_GetAppReportGenerator_YNABUnsupported(t *testing.T) {
	factory := DefaultAppReportGeneratorFactory{}

	generator, err := factory.GetAppReportGenerator(constants.YNAB)

	assert.Error(t, err)
	assert.Nil(t, generator)
}

func TestDefaultAppReportGeneratorFactory_ImplementsInterface(t *testing.T) {
	var _ AppReportGeneratorFactory = DefaultAppReportGeneratorFactory{}
}

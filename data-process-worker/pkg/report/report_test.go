package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/verasthiago/verancial/data-process-worker/pkg/report/firsttech"
	"github.com/verasthiago/verancial/data-process-worker/pkg/report/nubank"
	"github.com/verasthiago/verancial/data-process-worker/pkg/report/scotiabank"
	"github.com/verasthiago/verancial/data-process-worker/pkg/report/wise"
	"github.com/verasthiago/verancial/shared/constants"
)

func TestDefaultReportProcessorFactory_GetReportProcessor(t *testing.T) {
	factory := DefaultReportProcessorFactory{}

	tests := []struct {
		name   string
		bankId constants.BankId
		want   interface{}
	}{
		{name: "ScotiaBank", bankId: constants.ScotiaBank, want: scotiabank.ScotiaBankReportProcessor{}},
		{name: "ScotiaBankCC", bankId: constants.ScotiaBankCC, want: scotiabank.ScotiaBankCCReportProcessor{}},
		{name: "Nubank", bankId: constants.Nubank, want: nubank.NubankReportProcessor{}},
		{name: "Wise", bankId: constants.Wise, want: wise.WiseReportProcessor{}},
		{name: "FirstTech", bankId: constants.FirstTech, want: firsttech.FirstTechReportProcessor{}},
		{name: "FirstTechCC", bankId: constants.FirstTechCC, want: firsttech.FirstTechCCReportProcessor{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := factory.GetReportProcessor(tt.bankId)

			require.NoError(t, err)
			assert.IsType(t, tt.want, processor)
		})
	}

	t.Run("unsupported bank id returns error", func(t *testing.T) {
		processor, err := factory.GetReportProcessor(constants.BankId("not-a-real-bank"))

		require.Error(t, err)
		assert.Nil(t, processor)
		assert.EqualError(t, err, "bank not supported")
	})

	t.Run("empty bank id returns error", func(t *testing.T) {
		processor, err := factory.GetReportProcessor(constants.BankId(""))

		require.Error(t, err)
		assert.Nil(t, processor)
	})
}

package nubank

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	nubankmodel "github.com/verasthiago/verancial/data-process-worker/pkg/models/nubank"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

func writeCSV(t *testing.T, lines []string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "report.csv")
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func TestNubankReportProcessor_LoadFromCSV(t *testing.T) {
	t.Run("parses well formed rows (DD/MM/YYYY date, payee derived from description)", func(t *testing.T) {
		path := writeCSV(t, []string{
			`15/01/2024,-52.34,,Uber - UBER TRIP SAO PAULO`,
			`01/02/2024,1500.00,,Salary Deposit`,
		})

		processor := NubankReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		require.Len(t, result, 2)

		first, ok := result[0].(*nubankmodel.Nubank)
		require.True(t, ok)
		assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), first.Date)
		assert.Equal(t, float32(-52.34), first.Amount)
		assert.Equal(t, "Uber - UBER TRIP SAO PAULO", first.Description)
		// getPayee splits on " - " and takes the second part when present
		assert.Equal(t, "UBER TRIP SAO PAULO", first.Payee)

		second := result[1].(*nubankmodel.Nubank)
		assert.Equal(t, float32(1500.00), second.Amount)
		// No " - " separator: payee falls back to full description
		assert.Equal(t, "Salary Deposit", second.Payee)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		processor := NubankReportProcessor{}
		result, err := processor.LoadFromCSV(filepath.Join(t.TempDir(), "missing.csv"))

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed date returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			`2024-01-15,-52.34,,Bad Date Format`,
		})

		processor := NubankReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed amount returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			`15/01/2024,not-a-number,,Some Description`,
		})

		processor := NubankReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty file returns empty slice", func(t *testing.T) {
		path := writeCSV(t, []string{})

		processor := NubankReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestNubankReportProcessor_Process(t *testing.T) {
	payload := &types.ReportProcessQueuePayload{UserId: "user-1", BankId: "b5604a4f-1389-45ae-af18-915acc268fed"}

	t.Run("maps fields with BRL currency and no metadata", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&nubankmodel.Nubank{
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Amount:      -52.34,
				Description: "Uber - UBER TRIP SAO PAULO",
				Payee:       "UBER TRIP SAO PAULO",
			},
		}

		processor := NubankReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		tx := result[0]
		assert.Equal(t, "user-1", tx.UserId)
		assert.Equal(t, float32(-52.34), tx.Amount)
		assert.Equal(t, "UBER TRIP SAO PAULO", tx.Payee)
		assert.Equal(t, "Uber - UBER TRIP SAO PAULO", tx.Description)
		assert.Equal(t, "BRL", tx.Currency)
		assert.Equal(t, payload.BankId, tx.BankId)
		assert.Nil(t, tx.Metadata)
	})

	t.Run("excludes transactions on or before lastDbTransaction.Date (strict After)", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&nubankmodel.Nubank{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Amount: -1, Description: "same-day-excluded", Payee: "same"},
			&nubankmodel.Nubank{Date: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), Amount: -1, Description: "after", Payee: "after"},
		}

		processor := NubankReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "after", result[0].Description)
	})

	t.Run("returns error when bank transaction has wrong underlying type", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{"wrong-type"}

		processor := NubankReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetPayee(t *testing.T) {
	tests := []struct {
		name   string
		record string
		want   string
	}{
		{name: "splits on ' - ' and returns second part", record: "Uber - UBER TRIP", want: "UBER TRIP"},
		{name: "no separator returns whole string", record: "Salary Deposit", want: "Salary Deposit"},
		{name: "multiple separators returns second part only", record: "A - B - C", want: "B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getPayee(tt.record))
		})
	}
}

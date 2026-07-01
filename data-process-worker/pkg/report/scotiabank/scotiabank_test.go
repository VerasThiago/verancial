package scotiabank

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	scotiabankmodel "github.com/verasthiago/verancial/data-process-worker/pkg/models/scotiabank"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

// writeCSV writes lines (no trailing newline requirement) to a temp csv file
// and returns its path.
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

func TestScotiaBankReportProcessor_LoadFromCSV(t *testing.T) {
	t.Run("parses well formed rows", func(t *testing.T) {
		path := writeCSV(t, []string{
			`00000001,2024-01-15,GROCERY STORE PURCHASE,GROCERY STORE,,-52.34`,
			`00000001,2024-02-01,PAYROLL DEPOSIT,EMPLOYER INC,,1500.00`,
		})

		processor := ScotiaBankReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		require.Len(t, result, 2)

		first, ok := result[0].(*scotiabankmodel.ScotiaBank)
		require.True(t, ok)
		assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), first.Date)
		assert.Equal(t, float32(-52.34), first.Amount)
		assert.Equal(t, "GROCERY STORE PURCHASE", first.Description)
		assert.Equal(t, "GROCERY STORE", first.Payee)

		second, ok := result[1].(*scotiabankmodel.ScotiaBank)
		require.True(t, ok)
		assert.Equal(t, float32(1500.00), second.Amount)
	})

	t.Run("collapses repeated whitespace in description and payee", func(t *testing.T) {
		path := writeCSV(t, []string{
			`00000001,2024-01-15,GROCERY   STORE    PURCHASE,GROCERY  STORE  ,,-52.34`,
		})

		processor := ScotiaBankReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		require.Len(t, result, 1)
		first := result[0].(*scotiabankmodel.ScotiaBank)
		assert.Equal(t, "GROCERY STORE PURCHASE", first.Description)
		assert.Equal(t, "GROCERY STORE", first.Payee)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		processor := ScotiaBankReportProcessor{}
		result, err := processor.LoadFromCSV(filepath.Join(t.TempDir(), "does-not-exist.csv"))

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed date returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			`00000001,not-a-date,GROCERY STORE PURCHASE,GROCERY STORE,,-52.34`,
		})

		processor := ScotiaBankReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed amount returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			`00000001,2024-01-15,GROCERY STORE PURCHASE,GROCERY STORE,,not-a-number`,
		})

		processor := ScotiaBankReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty file returns empty slice", func(t *testing.T) {
		path := writeCSV(t, []string{})

		processor := ScotiaBankReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("inconsistent column count returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			`00000001,2024-01-15,GROCERY STORE PURCHASE,GROCERY STORE,,-52.34`,
			`too,few,columns`,
		})

		processor := ScotiaBankReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestScotiaBankReportProcessor_Process(t *testing.T) {
	payload := &types.ReportProcessQueuePayload{UserId: "user-1", BankId: "92f1aff9-03ae-4e6e-bde2-7799e849d181"}

	t.Run("maps fields and applies currency/bankid", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&scotiabankmodel.ScotiaBank{
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Amount:      -52.34,
				Description: "GROCERY STORE PURCHASE",
				Payee:       "GROCERY STORE",
			},
		}

		processor := ScotiaBankReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		tx := result[0]
		assert.Equal(t, "user-1", tx.UserId)
		assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), tx.Date)
		assert.Equal(t, float32(-52.34), tx.Amount)
		assert.Equal(t, "GROCERY STORE", tx.Payee)
		assert.Equal(t, "GROCERY STORE PURCHASE", tx.Description)
		assert.Equal(t, "CAD", tx.Currency)
		assert.Equal(t, payload.BankId, tx.BankId)
		assert.Nil(t, tx.Metadata)
	})

	t.Run("excludes transactions on or before lastDbTransaction.Date (strict After)", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&scotiabankmodel.ScotiaBank{Date: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), Amount: -1, Description: "before", Payee: "before"},
			&scotiabankmodel.ScotiaBank{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Amount: -1, Description: "same-day-excluded", Payee: "same"},
			&scotiabankmodel.ScotiaBank{Date: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), Amount: -1, Description: "after", Payee: "after"},
		}

		processor := ScotiaBankReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "after", result[0].Description)
	})

	t.Run("no transactions after lastDbTransaction returns empty result", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&scotiabankmodel.ScotiaBank{Date: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), Amount: -1, Description: "before", Payee: "before"},
		}

		processor := ScotiaBankReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("categorizes payee via category guesser", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&scotiabankmodel.ScotiaBank{
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Amount:      -20,
				Description: "FITNESS WORLD GEORGIA",
				Payee:       "FITNESS WORLD GEORGIA",
			},
		}

		processor := ScotiaBankReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "Gym", result[0].Category)
	})

	t.Run("returns error when bank transaction has wrong underlying type", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{"not-a-scotiabank-transaction"}

		processor := ScotiaBankReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

package firsttech

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	firsttechmodel "github.com/verasthiago/verancial/data-process-worker/pkg/models/firsttech"
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

// firsttech row layout (13 columns, index 0-12):
// 0 TransactionID, 1 PostingDate, 2 EffectiveDate, 3 TransactionType, 4 Amount,
// 5 CheckNumber, 6 ReferenceNumber, 7 Description, 8 TransactionCategory,
// 9 Type, 10 Balance, 11 Memo, 12 ExtendedDescription
const firstTechHeader = `Transaction ID,Posting Date,Effective Date,Transaction Type,Amount,Check Number,Reference Number,Description,Transaction Category,Type,Balance,Memo,Extended Description`

func sampleFirstTechRow(id, postingDate, effectiveDate, amount, description, balance string) string {
	return id + `,` + postingDate + `,` + effectiveDate + `,DEBIT,` + amount + `,,REF123,` + description + `,Shopping,POS,` + balance + `,,`
}

func TestFirstTechReportProcessor_LoadFromCSV(t *testing.T) {
	t.Run("skips the header row (detected via record[0]=='Transaction ID') and parses data rows", func(t *testing.T) {
		path := writeCSV(t, []string{
			firstTechHeader,
			sampleFirstTechRow("TX1", "1/15/2024", "1/16/2024", "-45.67", "GROCERY STORE", "1000.00"),
		})

		processor := FirstTechReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		require.Len(t, result, 1)

		first, ok := result[0].(*firsttechmodel.FirstTech)
		require.True(t, ok)
		assert.Equal(t, "TX1", first.TransactionID)
		assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), first.PostingDate)
		assert.Equal(t, time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), first.EffectiveDate)
		assert.Equal(t, "DEBIT", first.TransactionType)
		assert.Equal(t, float32(-45.67), first.Amount)
		assert.Equal(t, "REF123", first.ReferenceNumber)
		assert.Equal(t, "GROCERY STORE", first.Description)
		assert.Equal(t, "Shopping", first.TransactionCategory)
		assert.Equal(t, "POS", first.Type)
		assert.Equal(t, float32(1000.00), first.Balance)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		processor := FirstTechReportProcessor{}
		result, err := processor.LoadFromCSV(filepath.Join(t.TempDir(), "missing.csv"))

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed posting date returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			firstTechHeader,
			sampleFirstTechRow("TX1", "not-a-date", "1/16/2024", "-45.67", "GROCERY STORE", "1000.00"),
		})

		processor := FirstTechReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed effective date returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			firstTechHeader,
			sampleFirstTechRow("TX1", "1/15/2024", "not-a-date", "-45.67", "GROCERY STORE", "1000.00"),
		})

		processor := FirstTechReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed amount returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			firstTechHeader,
			sampleFirstTechRow("TX1", "1/15/2024", "1/16/2024", "not-a-number", "GROCERY STORE", "1000.00"),
		})

		processor := FirstTechReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed balance returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			firstTechHeader,
			sampleFirstTechRow("TX1", "1/15/2024", "1/16/2024", "-45.67", "GROCERY STORE", "not-a-number"),
		})

		processor := FirstTechReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("only header returns empty slice", func(t *testing.T) {
		path := writeCSV(t, []string{firstTechHeader})

		processor := FirstTechReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestFirstTechReportProcessor_Process(t *testing.T) {
	payload := &types.ReportProcessQueuePayload{UserId: "user-1", BankId: "91c1ae86-05e7-4b57-b288-cd6fe5a61ccb"}

	t.Run("maps fields using PostingDate as Date, USD currency, no metadata", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&firsttechmodel.FirstTech{
				TransactionID: "TX1",
				PostingDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				EffectiveDate: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
				Amount:        -45.67,
				Description:   "GROCERY STORE",
			},
		}

		processor := FirstTechReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		tx := result[0]
		assert.Equal(t, "user-1", tx.UserId)
		assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), tx.Date)
		assert.Equal(t, float32(-45.67), tx.Amount)
		assert.Equal(t, "GROCERY STORE", tx.Payee)
		assert.Equal(t, "GROCERY STORE", tx.Description)
		assert.Equal(t, "USD", tx.Currency)
		assert.Equal(t, payload.BankId, tx.BankId)
		assert.Nil(t, tx.Metadata)
	})

	t.Run("excludes transactions on or before lastDbTransaction.Date (strict After, uses PostingDate)", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&firsttechmodel.FirstTech{PostingDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Description: "same-day-excluded"},
			&firsttechmodel.FirstTech{PostingDate: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), Description: "after"},
		}

		processor := FirstTechReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "after", result[0].Description)
	})

	t.Run("returns error when bank transaction has wrong underlying type", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{[]int{1, 2, 3}}

		processor := FirstTechReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

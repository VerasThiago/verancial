package firsttech

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	firsttechmodel "github.com/verasthiago/verancial/data-process-worker/pkg/models/firsttech"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

func TestFirstTechCCReportProcessor_LoadFromCSV(t *testing.T) {
	t.Run("skips the header row and parses data rows with the same layout as the checking processor", func(t *testing.T) {
		path := writeCSV(t, []string{
			firstTechHeader,
			sampleFirstTechRow("TX1", "1/15/2024", "1/16/2024", "-120.00", "AMAZON.COM PURCHASE", "500.00"),
		})

		processor := FirstTechCCReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		require.Len(t, result, 1)

		first, ok := result[0].(*firsttechmodel.FirstTech)
		require.True(t, ok)
		assert.Equal(t, "TX1", first.TransactionID)
		assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), first.PostingDate)
		assert.Equal(t, float32(-120.00), first.Amount)
		assert.Equal(t, "AMAZON.COM PURCHASE", first.Description)
		assert.Equal(t, float32(500.00), first.Balance)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		processor := FirstTechCCReportProcessor{}
		result, err := processor.LoadFromCSV(filepath.Join(t.TempDir(), "missing.csv"))

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed amount returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			firstTechHeader,
			sampleFirstTechRow("TX1", "1/15/2024", "1/16/2024", "not-a-number", "AMAZON.COM PURCHASE", "500.00"),
		})

		processor := FirstTechCCReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed posting date returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			firstTechHeader,
			sampleFirstTechRow("TX1", "not-a-date", "1/16/2024", "-120.00", "AMAZON.COM PURCHASE", "500.00"),
		})

		processor := FirstTechCCReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed effective date returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			firstTechHeader,
			sampleFirstTechRow("TX1", "1/15/2024", "not-a-date", "-120.00", "AMAZON.COM PURCHASE", "500.00"),
		})

		processor := FirstTechCCReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed balance returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			firstTechHeader,
			sampleFirstTechRow("TX1", "1/15/2024", "1/16/2024", "-120.00", "AMAZON.COM PURCHASE", "not-a-number"),
		})

		processor := FirstTechCCReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("only header returns empty slice", func(t *testing.T) {
		path := writeCSV(t, []string{firstTechHeader})

		processor := FirstTechCCReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestFirstTechCCReportProcessor_Process(t *testing.T) {
	payload := &types.ReportProcessQueuePayload{UserId: "user-1", BankId: "a317dc85-81bb-40ad-a12f-d8b546c1230b"}

	t.Run("maps fields using PostingDate as Date, USD currency, no sign flip, no metadata", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&firsttechmodel.FirstTech{
				TransactionID: "TX1",
				PostingDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				EffectiveDate: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
				Amount:        -120.00,
				Description:   "AMAZON.COM PURCHASE",
			},
		}

		processor := FirstTechCCReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		tx := result[0]
		assert.Equal(t, "user-1", tx.UserId)
		assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), tx.Date)
		// Credit card variant does not flip sign - amount is passed through unchanged
		assert.Equal(t, float32(-120.00), tx.Amount)
		assert.Equal(t, "AMAZON.COM PURCHASE", tx.Payee)
		assert.Equal(t, "USD", tx.Currency)
		assert.Equal(t, payload.BankId, tx.BankId)
		assert.Nil(t, tx.Metadata)
	})

	t.Run("excludes transactions on or before lastDbTransaction.Date (strict After)", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&firsttechmodel.FirstTech{PostingDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Description: "same-day-excluded"},
			&firsttechmodel.FirstTech{PostingDate: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), Description: "after"},
		}

		processor := FirstTechCCReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "after", result[0].Description)
	})

	t.Run("returns error when bank transaction has wrong underlying type", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{true}

		processor := FirstTechCCReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

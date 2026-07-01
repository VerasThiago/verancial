package wise

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wisemodel "github.com/verasthiago/verancial/data-process-worker/pkg/models/wise"
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

// header matching the real Wise export column layout (18 columns, index 0-17):
// 0 TransactionID, 1 Status, 2 Direction, 3 Created on, 4 Finished on,
// 5 Source fee amount, 6 Source fee currency, 7 Target fee amount, 8 Target fee currency,
// 9 Source name, 10 Source amount (after fees), 11 Source currency,
// 12 Target name, 13 Target amount (after fees), 14 Target currency,
// 15 Exchange rate, 16 Reference, 17 Batch
const wiseHeader = `ID,Status,Direction,Created on,Finished on,Source fee amount,Source fee currency,Target fee amount,Target fee currency,Source name,Source amount (after fees),Source currency,Target name,Target amount (after fees),Target currency,Exchange rate,Reference,Batch`

func sampleWiseRow(id, direction, createdOn, finishedOn, targetName, targetAmount, reference string) string {
	return id + `,COMPLETED,` + direction + `,` + createdOn + `,` + finishedOn + `,1.50,CAD,0.00,USD,John Doe,100.00,CAD,` + targetName + `,` + targetAmount + `,USD,0.75,` + reference + `,batch-1`
}

func TestWiseReportProcessor_LoadFromCSV(t *testing.T) {
	t.Run("skips header row and parses well formed rows", func(t *testing.T) {
		path := writeCSV(t, []string{
			wiseHeader,
			sampleWiseRow("TX1", "OUT", "2024-01-15 10:30:00", "2024-01-15 10:35:00", "Jane Smith", "-98.50", "Invoice 123"),
		})

		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		require.Len(t, result, 1)

		first, ok := result[0].(*wisemodel.Wise)
		require.True(t, ok)
		assert.Equal(t, "TX1", first.TransactionID)
		assert.Equal(t, "COMPLETED", first.Status)
		assert.Equal(t, "OUT", first.Direction)
		assert.Equal(t, time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), first.CreatedOn)
		assert.Equal(t, time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC), first.FinishedOn)
		assert.Equal(t, float32(1.50), first.SourceFeeAmount)
		assert.Equal(t, "CAD", first.SourceFeeCurrency)
		assert.Equal(t, float32(0.00), first.TargetFeeAmount)
		assert.Equal(t, "USD", first.TargetFeeCurrency)
		assert.Equal(t, "John Doe", first.SourceName)
		assert.Equal(t, float32(100.00), first.SourceAmountAfterFees)
		assert.Equal(t, "CAD", first.SourceCurrency)
		assert.Equal(t, "Jane Smith", first.TargetName)
		assert.Equal(t, float32(-98.50), first.TargetAmountAfterFees)
		assert.Equal(t, "USD", first.TargetCurrency)
		assert.Equal(t, float32(0.75), first.ExchangeRate)
		assert.Equal(t, "Invoice 123", first.Reference)
		assert.Equal(t, "batch-1", first.Batch)
	})

	t.Run("only header returns empty slice", func(t *testing.T) {
		path := writeCSV(t, []string{wiseHeader})

		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty file returns error reading header", func(t *testing.T) {
		path := writeCSV(t, []string{})

		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(filepath.Join(t.TempDir(), "missing.csv"))

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed CreatedOn date returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			wiseHeader,
			sampleWiseRow("TX1", "OUT", "not-a-date", "2024-01-15 10:35:00", "Jane Smith", "-98.50", "Invoice 123"),
		})

		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed FinishedOn date returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			wiseHeader,
			sampleWiseRow("TX1", "OUT", "2024-01-15 10:30:00", "not-a-date", "Jane Smith", "-98.50", "Invoice 123"),
		})

		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed source fee amount returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			wiseHeader,
			`TX1,COMPLETED,OUT,2024-01-15 10:30:00,2024-01-15 10:35:00,not-a-number,CAD,0.00,USD,John Doe,100.00,CAD,Jane Smith,-98.50,USD,0.75,Invoice 123,batch-1`,
		})

		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed target fee amount returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			wiseHeader,
			`TX1,COMPLETED,OUT,2024-01-15 10:30:00,2024-01-15 10:35:00,1.50,CAD,not-a-number,USD,John Doe,100.00,CAD,Jane Smith,-98.50,USD,0.75,Invoice 123,batch-1`,
		})

		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed source amount after fees returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			wiseHeader,
			`TX1,COMPLETED,OUT,2024-01-15 10:30:00,2024-01-15 10:35:00,1.50,CAD,0.00,USD,John Doe,not-a-number,CAD,Jane Smith,-98.50,USD,0.75,Invoice 123,batch-1`,
		})

		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed target amount after fees returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			wiseHeader,
			`TX1,COMPLETED,OUT,2024-01-15 10:30:00,2024-01-15 10:35:00,1.50,CAD,0.00,USD,John Doe,100.00,CAD,Jane Smith,not-a-number,USD,0.75,Invoice 123,batch-1`,
		})

		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed exchange rate returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			wiseHeader,
			`TX1,COMPLETED,OUT,2024-01-15 10:30:00,2024-01-15 10:35:00,1.50,CAD,0.00,USD,John Doe,100.00,CAD,Jane Smith,-98.50,USD,not-a-number,Invoice 123,batch-1`,
		})

		processor := WiseReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestWiseReportProcessor_Process(t *testing.T) {
	payload := &types.ReportProcessQueuePayload{UserId: "user-1", BankId: "5ae0c7ff-20c2-4bb8-af55-35347df9a9fd"}

	t.Run("maps fields using FinishedOn as Date and TargetAmountAfterFees as Amount, with full metadata", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&wisemodel.Wise{
				TransactionID:         "TX1",
				Status:                "COMPLETED",
				Direction:             "OUT",
				CreatedOn:             time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				FinishedOn:            time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC),
				SourceFeeAmount:       1.5,
				SourceFeeCurrency:     "CAD",
				TargetFeeAmount:       0,
				TargetFeeCurrency:     "USD",
				SourceName:            "John Doe",
				SourceAmountAfterFees: 100,
				SourceCurrency:        "CAD",
				TargetName:            "Jane Smith",
				TargetAmountAfterFees: -98.5,
				TargetCurrency:        "USD",
				ExchangeRate:          0.75,
				Reference:             "Invoice 123",
				Batch:                 "batch-1",
			},
		}

		processor := WiseReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		tx := result[0]
		assert.Equal(t, "user-1", tx.UserId)
		// Date comes from FinishedOn, not CreatedOn
		assert.Equal(t, time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC), tx.Date)
		assert.Equal(t, float32(-98.5), tx.Amount)
		assert.Equal(t, "Jane Smith", tx.Payee)
		assert.Equal(t, "Invoice 123", tx.Description)
		assert.Equal(t, "CAD", tx.Currency)
		assert.Equal(t, payload.BankId, tx.BankId)
		require.NotNil(t, tx.Metadata)
		assert.Equal(t, "TX1", tx.Metadata["TransactionID"])
		assert.Equal(t, "COMPLETED", tx.Metadata["Status"])
		assert.Equal(t, "OUT", tx.Metadata["Direction"])
		assert.Equal(t, "CAD", tx.Metadata["SourceFeeCurrency"])
		assert.Equal(t, "USD", tx.Metadata["TargetFeeCurrency"])
		assert.Equal(t, "John Doe", tx.Metadata["SourceName"])
		assert.Equal(t, "CAD", tx.Metadata["SourceCurrency"])
		assert.Equal(t, "USD", tx.Metadata["TargetCurrency"])
		assert.Equal(t, "batch-1", tx.Metadata["Batch"])
	})

	t.Run("comparison against lastDbTransaction uses CreatedOn, not FinishedOn", func(t *testing.T) {
		// lastDbTransaction.Date sits between CreatedOn and FinishedOn to prove which field drives inclusion.
		lastDbTransaction := &models.Transaction{Date: time.Date(2024, 1, 15, 10, 32, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&wisemodel.Wise{
				TransactionID: "TX1",
				CreatedOn:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), // before lastDbTransaction -> excluded
				FinishedOn:    time.Date(2024, 1, 15, 10, 40, 0, 0, time.UTC), // after lastDbTransaction
				TargetName:    "included-if-finishedon-used",
				Reference:     "ref",
			},
		}

		processor := WiseReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		assert.Empty(t, result, "CreatedOn is before lastDbTransaction.Date, so it must be excluded regardless of FinishedOn")
	})

	t.Run("returns error when bank transaction has wrong underlying type", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{123}

		processor := WiseReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestParseFloatToString(t *testing.T) {
	tests := []struct {
		name   string
		number float32
		want   string
	}{
		{name: "positive value", number: 1.5, want: "1.5"},
		{name: "zero", number: 0, want: "0"},
		{name: "negative value", number: -98.5, want: "-98.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ParseFloatToString(tt.number))
		})
	}
}

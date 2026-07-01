package scotiabank

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	scotiabankmodel "github.com/verasthiago/verancial/data-process-worker/pkg/models/scotiabank"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

func TestScotiaBankCCReportProcessor_LoadFromCSV(t *testing.T) {
	t.Run("parses well formed rows with cc column layout (record[6]=amount, payee/description swapped)", func(t *testing.T) {
		path := writeCSV(t, []string{
			`00000001,2024-01-15,AMAZON.COM,ONLINE PURCHASE FROM AMAZON,,,-99.99`,
			`00000001,2024-02-01,PAYMENT RECEIVED,THANK YOU FOR YOUR PAYMENT,,,250.00`,
		})

		processor := ScotiaBankCCReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		require.Len(t, result, 2)

		first, ok := result[0].(*scotiabankmodel.ScotiaBank)
		require.True(t, ok)
		assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), first.Date)
		assert.Equal(t, float32(-99.99), first.Amount)
		// CC variant: record[2]=payee, record[3]=description (swapped vs regular scotiabank)
		assert.Equal(t, "AMAZON.COM", first.Payee)
		assert.Equal(t, "ONLINE PURCHASE FROM AMAZON", first.Description)

		second := result[1].(*scotiabankmodel.ScotiaBank)
		assert.Equal(t, float32(250.00), second.Amount)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		processor := ScotiaBankCCReportProcessor{}
		result, err := processor.LoadFromCSV(filepath.Join(t.TempDir(), "missing.csv"))

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("malformed amount column returns error", func(t *testing.T) {
		path := writeCSV(t, []string{
			`00000001,2024-01-15,AMAZON.COM,ONLINE PURCHASE,,,not-a-number`,
		})

		processor := ScotiaBankCCReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty file returns empty slice", func(t *testing.T) {
		path := writeCSV(t, []string{})

		processor := ScotiaBankCCReportProcessor{}
		result, err := processor.LoadFromCSV(path)

		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestScotiaBankCCReportProcessor_Process(t *testing.T) {
	payload := &types.ReportProcessQueuePayload{UserId: "user-1", BankId: "8462037f-7615-406a-b8fc-214105becd33"}

	t.Run("maps fields with CAD currency and no metadata", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&scotiabankmodel.ScotiaBank{
				Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Amount:      -99.99,
				Description: "ONLINE PURCHASE FROM AMAZON",
				Payee:       "AMAZON.COM",
			},
		}

		processor := ScotiaBankCCReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		tx := result[0]
		assert.Equal(t, "user-1", tx.UserId)
		assert.Equal(t, float32(-99.99), tx.Amount)
		assert.Equal(t, "AMAZON.COM", tx.Payee)
		assert.Equal(t, "ONLINE PURCHASE FROM AMAZON", tx.Description)
		assert.Equal(t, "CAD", tx.Currency)
		assert.Equal(t, payload.BankId, tx.BankId)
		assert.Nil(t, tx.Metadata)
	})

	t.Run("excludes transactions on or before lastDbTransaction.Date (strict After)", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{
			&scotiabankmodel.ScotiaBank{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Amount: -1, Description: "same-day-excluded", Payee: "same"},
			&scotiabankmodel.ScotiaBank{Date: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), Amount: -1, Description: "after", Payee: "after"},
		}

		processor := ScotiaBankCCReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "after", result[0].Description)
	})

	t.Run("returns error when bank transaction has wrong underlying type", func(t *testing.T) {
		lastDbTransaction := &models.Transaction{Date: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{42}

		processor := ScotiaBankCCReportProcessor{}
		result, err := processor.Process(bankTransactions, payload, lastDbTransaction)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

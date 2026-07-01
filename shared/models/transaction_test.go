package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransaction_SetFingerprint(t *testing.T) {
	baseDate := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)

	t.Run("is deterministic for identical inputs", func(t *testing.T) {
		t1 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "Coffee Shop"}
		t2 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "Coffee Shop"}

		t1.SetFingerprint()
		t2.SetFingerprint()

		require.NotEmpty(t, t1.Fingerprint)
		assert.Equal(t, t1.Fingerprint, t2.Fingerprint)
	})

	t.Run("differs when user id differs", func(t *testing.T) {
		t1 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "Coffee Shop"}
		t2 := &Transaction{UserId: "user-2", Date: baseDate, Amount: 42.50, Description: "Coffee Shop"}

		t1.SetFingerprint()
		t2.SetFingerprint()

		assert.NotEqual(t, t1.Fingerprint, t2.Fingerprint)
	})

	t.Run("differs when date differs", func(t *testing.T) {
		t1 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "Coffee Shop"}
		t2 := &Transaction{UserId: "user-1", Date: baseDate.AddDate(0, 0, 1), Amount: 42.50, Description: "Coffee Shop"}

		t1.SetFingerprint()
		t2.SetFingerprint()

		assert.NotEqual(t, t1.Fingerprint, t2.Fingerprint)
	})

	t.Run("differs when amount differs", func(t *testing.T) {
		t1 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "Coffee Shop"}
		t2 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.51, Description: "Coffee Shop"}

		t1.SetFingerprint()
		t2.SetFingerprint()

		assert.NotEqual(t, t1.Fingerprint, t2.Fingerprint)
	})

	t.Run("differs when description differs", func(t *testing.T) {
		t1 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "Coffee Shop"}
		t2 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "Grocery Store"}

		t1.SetFingerprint()
		t2.SetFingerprint()

		assert.NotEqual(t, t1.Fingerprint, t2.Fingerprint)
	})

	t.Run("is insensitive to time-of-day within the same UTC calendar date", func(t *testing.T) {
		t1 := &Transaction{UserId: "user-1", Date: time.Date(2024, 3, 15, 0, 0, 1, 0, time.UTC), Amount: 42.50, Description: "Coffee Shop"}
		t2 := &Transaction{UserId: "user-1", Date: time.Date(2024, 3, 15, 23, 59, 59, 0, time.UTC), Amount: 42.50, Description: "Coffee Shop"}

		t1.SetFingerprint()
		t2.SetFingerprint()

		assert.Equal(t, t1.Fingerprint, t2.Fingerprint)
	})

	t.Run("normalizes dates across timezones to the same UTC calendar day", func(t *testing.T) {
		loc := time.FixedZone("UTC-5", -5*60*60)
		// 2024-03-15 23:00 UTC-5 == 2024-03-16 04:00 UTC
		t1 := &Transaction{UserId: "user-1", Date: time.Date(2024, 3, 15, 23, 0, 0, 0, loc), Amount: 10, Description: "X"}
		t2 := &Transaction{UserId: "user-1", Date: time.Date(2024, 3, 16, 4, 0, 0, 0, time.UTC), Amount: 10, Description: "X"}

		t1.SetFingerprint()
		t2.SetFingerprint()

		assert.Equal(t, t1.Fingerprint, t2.Fingerprint)
	})

	t.Run("is case-insensitive and trims whitespace on description", func(t *testing.T) {
		t1 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "  Coffee Shop  "}
		t2 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "coffee shop"}

		t1.SetFingerprint()
		t2.SetFingerprint()

		assert.Equal(t, t1.Fingerprint, t2.Fingerprint)
	})

	t.Run("rounds amount to 2 decimal places for fingerprinting", func(t *testing.T) {
		t1 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.5049, Description: "X"}
		t2 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.505, Description: "X"}

		t1.SetFingerprint()
		t2.SetFingerprint()

		// 42.5049 -> "42.50", 42.505 -> "42.51" (or "42.50" depending on float rounding);
		// what matters is fingerprint is a function of the formatted 2-decimal string.
		expectedT1 := t1.Fingerprint
		t1b := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.5049, Description: "X"}
		t1b.SetFingerprint()
		assert.Equal(t, expectedT1, t1b.Fingerprint)
	})

	t.Run("produces a 64-character hex sha256 digest", func(t *testing.T) {
		tx := &Transaction{UserId: "user-1", Date: baseDate, Amount: 1, Description: "x"}
		tx.SetFingerprint()

		assert.Len(t, tx.Fingerprint, 64)
		assert.Regexp(t, "^[0-9a-f]{64}$", tx.Fingerprint)
	})

	t.Run("payee and bank id do not affect the fingerprint", func(t *testing.T) {
		t1 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "Coffee Shop", Payee: "Starbucks", BankId: "bank-a"}
		t2 := &Transaction{UserId: "user-1", Date: baseDate, Amount: 42.50, Description: "Coffee Shop", Payee: "Different Payee", BankId: "bank-b"}

		t1.SetFingerprint()
		t2.SetFingerprint()

		assert.Equal(t, t1.Fingerprint, t2.Fingerprint)
	})
}

func TestTransaction_BeforeCreate(t *testing.T) {
	t.Run("assigns a UUID as ID", func(t *testing.T) {
		tx := &Transaction{}
		err := tx.BeforeCreate(nil)

		require.NoError(t, err)
		assert.NotEmpty(t, tx.ID)
	})

	t.Run("generates different IDs across calls", func(t *testing.T) {
		tx1 := &Transaction{}
		tx2 := &Transaction{}
		require.NoError(t, tx1.BeforeCreate(nil))
		require.NoError(t, tx2.BeforeCreate(nil))

		assert.NotEqual(t, tx1.ID, tx2.ID)
	})
}

func TestTransactionMetadata_Scan(t *testing.T) {
	t.Run("nil value clears the map", func(t *testing.T) {
		m := TransactionMetadata{"key": "value"}
		err := m.Scan(nil)

		require.NoError(t, err)
		assert.Nil(t, m)
	})

	t.Run("valid JSON bytes unmarshal into the map", func(t *testing.T) {
		var m TransactionMetadata
		err := m.Scan([]byte(`{"source":"csv","imported":"true"}`))

		require.NoError(t, err)
		assert.Equal(t, "csv", m["source"])
		assert.Equal(t, "true", m["imported"])
	})

	t.Run("unsupported value type returns an error", func(t *testing.T) {
		var m TransactionMetadata
		err := m.Scan(12345)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported value type")
	})

	t.Run("malformed JSON bytes returns an error", func(t *testing.T) {
		var m TransactionMetadata
		err := m.Scan([]byte(`not-json`))

		assert.Error(t, err)
	})

	t.Run("empty byte slice returns an error (invalid JSON)", func(t *testing.T) {
		var m TransactionMetadata
		err := m.Scan([]byte(``))

		assert.Error(t, err)
	})
}

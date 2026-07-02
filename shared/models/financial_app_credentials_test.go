package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/verasthiago/verancial/shared/constants"
)

func TestFinancialAppCredentialsMap_Scan(t *testing.T) {
	t.Run("nil value clears the map", func(t *testing.T) {
		f := FinancialAppCredentialsMap{constants.BudgetBakers: &FinancialAppCredentials{Login: "a"}}
		require.NoError(t, f.Scan(nil))
		assert.Nil(t, f)
	})

	t.Run("valid JSON bytes populate the map", func(t *testing.T) {
		var f FinancialAppCredentialsMap
		raw := []byte(`{"` + string(constants.BudgetBakers) + `":{"login":"user@example.com","password":"hunter2"}}`)

		require.NoError(t, f.Scan(raw))

		require.Contains(t, f, constants.BudgetBakers)
		cred := f[constants.BudgetBakers]
		assert.Equal(t, "user@example.com", cred.Login)
		assert.Equal(t, "hunter2", cred.Password)
	})

	t.Run("non-byte-slice value returns an error", func(t *testing.T) {
		var f FinancialAppCredentialsMap
		err := f.Scan(42)
		assert.Error(t, err)
	})

	t.Run("malformed JSON returns an error", func(t *testing.T) {
		var f FinancialAppCredentialsMap
		err := f.Scan([]byte(`not-json`))
		assert.Error(t, err)
	})
}

func TestFinancialAppCredentialsMap_Value(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		var f FinancialAppCredentialsMap
		v, err := f.Value()
		require.NoError(t, err)
		assert.Nil(t, v)
	})

	t.Run("round-trips through Scan", func(t *testing.T) {
		f := FinancialAppCredentialsMap{constants.BudgetBakers: &FinancialAppCredentials{Login: "user@example.com", Password: "hunter2"}}

		v, err := f.Value()
		require.NoError(t, err)

		var roundTripped FinancialAppCredentialsMap
		require.NoError(t, roundTripped.Scan(v))
		assert.Equal(t, f, roundTripped)
	})
}

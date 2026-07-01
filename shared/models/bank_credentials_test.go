package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/verasthiago/verancial/shared/constants"
)

func TestBankCredentialsMap_Scan(t *testing.T) {
	t.Run("nil value clears the map", func(t *testing.T) {
		b := BankCredentialsMap{constants.BudgetBakers: &BankCredentials{Login: "a"}}
		require.NoError(t, b.Scan(nil))
		assert.Nil(t, b)
	})

	t.Run("valid JSON bytes populate the map", func(t *testing.T) {
		var b BankCredentialsMap
		raw := []byte(`{"` + string(constants.BudgetBakers) + `":{"login":"user@example.com","password":"hunter2"}}`)

		require.NoError(t, b.Scan(raw))

		require.Contains(t, b, constants.BudgetBakers)
		cred := b[constants.BudgetBakers]
		assert.Equal(t, "user@example.com", cred.Login)
		assert.Equal(t, "hunter2", cred.Password)
	})

	t.Run("non-byte-slice value returns an error", func(t *testing.T) {
		var b BankCredentialsMap
		err := b.Scan(42)
		assert.Error(t, err)
	})

	t.Run("malformed JSON returns an error", func(t *testing.T) {
		var b BankCredentialsMap
		err := b.Scan([]byte(`not-json`))
		assert.Error(t, err)
	})
}

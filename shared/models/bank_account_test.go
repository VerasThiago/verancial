package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBankAccount_TableName(t *testing.T) {
	assert.Equal(t, "bank_accounts", (&BankAccount{}).TableName())
}

func TestUserBankAccount_TableName(t *testing.T) {
	assert.Equal(t, "user_bank_accounts", (&UserBankAccount{}).TableName())
}

func TestUserBankAccount_BeforeCreate(t *testing.T) {
	t.Run("assigns a UUID", func(t *testing.T) {
		uba := &UserBankAccount{}
		err := uba.BeforeCreate(nil)

		require.NoError(t, err)
		assert.NotEmpty(t, uba.ID)
	})

	t.Run("generates a different ID on each call", func(t *testing.T) {
		uba1 := &UserBankAccount{}
		uba2 := &UserBankAccount{}
		require.NoError(t, uba1.BeforeCreate(nil))
		require.NoError(t, uba2.BeforeCreate(nil))

		assert.NotEqual(t, uba1.ID, uba2.ID)
	})
}

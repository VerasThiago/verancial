package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppReport_IsSliceOfStringSlices(t *testing.T) {
	var report AppReport

	assert.Nil(t, report)

	report = AppReport{
		{"Date", "Amount", "Note", "Payee", "Currency"},
		{"2024-03-15", "42.500000", "Shopping | Online purchase", "Amazon", "USD"},
	}

	assert.Len(t, report, 2)
	assert.Equal(t, "Date", report[0][0])
	assert.Equal(t, "USD", report[1][4])
}

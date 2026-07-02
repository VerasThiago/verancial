package budgetbakers

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/verasthiago/verancial/shared/models"
)

func TestBudgetBakers_Generate_HeaderOnly(t *testing.T) {
	b := BudgetBakers{}

	report, err := b.Generate(nil)

	require.NoError(t, err)
	require.Len(t, report, 1)
	assert.Equal(t, []string{"Date", "Amount", "Note", "Payee", "Currency"}, report[0])
}

func TestBudgetBakers_Generate_FormatsTransactions(t *testing.T) {
	b := BudgetBakers{}

	date1 := time.Date(2024, time.March, 15, 10, 30, 0, 0, time.UTC)
	date2 := time.Date(2023, time.December, 1, 23, 59, 59, 0, time.FixedZone("EST", -5*60*60))

	transactions := []*models.Transaction{
		{
			Date:        date1,
			Amount:      42.5,
			Payee:       "Amazon",
			Description: "Online purchase",
			Category:    "Shopping",
			Currency:    "USD",
		},
		{
			Date:        date2,
			Amount:      -12.333333,
			Payee:       "Starbucks",
			Description: "Coffee",
			Category:    "Food & Drink",
			Currency:    "CAD",
		},
	}

	report, err := b.Generate(transactions)

	require.NoError(t, err)
	require.Len(t, report, 3)

	// Header
	assert.Equal(t, []string{"Date", "Amount", "Note", "Payee", "Currency"}, report[0])

	// Row 1: date formatted as YYYY-MM-DD in UTC, amount as %f (6 decimals), note "Category | Description"
	assert.Equal(t, "2024-03-15", report[1][0])
	assert.Equal(t, "42.500000", report[1][1])
	assert.Equal(t, "Shopping | Online purchase", report[1][2])
	assert.Equal(t, "Amazon", report[1][3])
	assert.Equal(t, "USD", report[1][4])

	// Row 2: date2 is 23:59:59 EST (UTC-5) on Dec 1 -> Dec 2 04:59:59 UTC
	assert.Equal(t, "2023-12-02", report[2][0])
	assert.Equal(t, "-12.333333", report[2][1])
	assert.Equal(t, "Food & Drink | Coffee", report[2][2])
	assert.Equal(t, "Starbucks", report[2][3])
	assert.Equal(t, "CAD", report[2][4])
}

func TestBudgetBakers_Generate_EmptyTransactions(t *testing.T) {
	b := BudgetBakers{}

	report, err := b.Generate([]*models.Transaction{})

	require.NoError(t, err)
	require.Len(t, report, 1)
	assert.Equal(t, []string{"Date", "Amount", "Note", "Payee", "Currency"}, report[0])
}

func TestBudgetBakers_Generate_ZeroAmount(t *testing.T) {
	b := BudgetBakers{}

	transactions := []*models.Transaction{
		{
			Date:        time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			Amount:      0,
			Payee:       "",
			Description: "",
			Category:    "",
			Currency:    "",
		},
	}

	report, err := b.Generate(transactions)

	require.NoError(t, err)
	require.Len(t, report, 2)
	assert.Equal(t, "0.000000", report[1][1])
	assert.Equal(t, " | ", report[1][2])
}

func chdirTemp(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	return tempDir
}

func TestBudgetBakers_Submit_WritesCSVFile(t *testing.T) {
	tempDir := chdirTemp(t)

	b := BudgetBakers{}
	user := &models.User{ID: "user-1"}

	appReport := [][]string{
		{"Date", "Amount", "Note", "Payee", "Currency"},
		{"2024-03-15", "42.500000", "Shopping | Online purchase", "Amazon", "USD"},
		{"2024-03-16", "-12.333333", "Food & Drink | Coffee", "Starbucks", "CAD"},
	}

	err := b.Submit(user, appReport)
	require.NoError(t, err)

	// Per helper.GetFileNameFromAppReport: reportSize=3, so
	// firstDate = appReport[2][0] = "2024-03-16", lastDate = appReport[1][0] = "2024-03-15"
	expectedFileName := "2024-03-16_to_2024-03-15.csv"

	fullPath := filepath.Join(tempDir, expectedFileName)
	_, statErr := os.Stat(fullPath)
	require.NoError(t, statErr, "expected file %s to exist", fullPath)

	f, err := os.Open(fullPath)
	require.NoError(t, err)
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	require.NoError(t, err)

	assert.Equal(t, appReport, rows)
}

func TestBudgetBakers_Submit_SingleRowReport(t *testing.T) {
	chdirTemp(t)

	b := BudgetBakers{}
	user := &models.User{ID: "user-1"}

	appReport := [][]string{
		{"Date", "Amount", "Note", "Payee", "Currency"},
	}

	err := b.Submit(user, appReport)
	require.NoError(t, err)

	// reportSize == 1 -> filename is appReport[0][0] == "Date"
	_, statErr := os.Stat("Date")
	require.NoError(t, statErr)

	f, err := os.Open("Date")
	require.NoError(t, err)
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, appReport, rows)
}

func TestBudgetBakers_Submit_EmptyReport(t *testing.T) {
	chdirTemp(t)

	b := BudgetBakers{}
	user := &models.User{ID: "user-1"}

	appReport := [][]string{}

	err := b.Submit(user, appReport)
	require.NoError(t, err)

	// reportSize == 0 -> filename is "empty"
	_, statErr := os.Stat("empty")
	require.NoError(t, statErr)
}

func TestBudgetBakers_Submit_InvalidDirectoryReturnsError(t *testing.T) {
	// Use a path with a directory component that does not exist so os.Create fails.
	chdirTemp(t)

	b := BudgetBakers{}
	user := &models.User{ID: "user-1"}

	appReport := [][]string{
		{"nonexistent-dir/report.csv"},
	}

	err := b.Submit(user, appReport)
	assert.Error(t, err)
}

func TestBudgetBakers_GetLastTransaction_ValidDate(t *testing.T) {
	b := BudgetBakers{}

	result, err := b.GetLastTransaction(nil, "bank-1", "January 2 2024")

	require.NoError(t, err)
	assert.Equal(t, 2024, result.Year())
	assert.Equal(t, time.January, result.Month())
	assert.Equal(t, 2, result.Day())
	assert.True(t, result.Equal(time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)))
}

func TestBudgetBakers_GetLastTransaction_InvalidDate(t *testing.T) {
	b := BudgetBakers{}

	tests := []struct {
		name string
		date string
	}{
		{"garbage string", "not-a-date"},
		{"wrong format ISO", "2024-01-02"},
		{"empty string", ""},
		{"missing year", "January 2"},
		{"day out of range", "January 45 2024"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := b.GetLastTransaction(nil, "bank-1", tt.date)
			assert.Error(t, err)
			assert.True(t, result.IsZero())
		})
	}
}

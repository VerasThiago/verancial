package helper

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/types"
)

func TestGetPythonScriptsPath(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)

	got, err := GetPythonScriptsPath()

	require.NoError(t, err)
	assert.Equal(t, path.Join(pwd, PYTHON_SCRIPTS_PATH), got)
}

func TestGetFileNameFromAppReport(t *testing.T) {
	tests := []struct {
		name     string
		report   types.AppReport
		expected string
	}{
		{
			name:     "empty report",
			report:   types.AppReport{},
			expected: "empty",
		},
		{
			name:     "nil report",
			report:   nil,
			expected: "empty",
		},
		{
			name: "single row report returns first cell of that row",
			report: types.AppReport{
				{"Date", "Amount", "Note", "Payee", "Currency"},
			},
			expected: "Date",
		},
		{
			name: "two rows: header + one transaction",
			report: types.AppReport{
				{"Date", "Amount", "Note", "Payee", "Currency"},
				{"2024-03-15", "42.500000", "Shopping | Online purchase", "Amazon", "USD"},
			},
			// reportSize=2: firstDate = report[1][0] = "2024-03-15", lastDate = report[1][0] = "2024-03-15"
			expected: "2024-03-15_to_2024-03-15.csv",
		},
		{
			name: "three rows: header + two transactions",
			report: types.AppReport{
				{"Date", "Amount", "Note", "Payee", "Currency"},
				{"2024-03-15", "42.500000", "Shopping | Online purchase", "Amazon", "USD"},
				{"2024-03-20", "-12.333333", "Food & Drink | Coffee", "Starbucks", "CAD"},
			},
			// reportSize=3: firstDate = report[2][0] = "2024-03-20", lastDate = report[1][0] = "2024-03-15"
			expected: "2024-03-20_to_2024-03-15.csv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFileNameFromAppReport(tt.report)
			assert.Equal(t, tt.expected, got)
		})
	}
}

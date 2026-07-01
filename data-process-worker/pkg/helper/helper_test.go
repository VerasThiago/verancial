package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAmountFloat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float32
		wantErr bool
	}{
		{name: "plain positive number", input: "52.34", want: 52.34},
		{name: "plain negative number", input: "-52.34", want: -52.34},
		{name: "dollar sign is stripped", input: "$52.34", want: 52.34},
		{name: "comma thousands separator is stripped", input: "1,500.00", want: 1500.00},
		{name: "dollar sign and comma combined", input: "$1,500.00", want: 1500.00},
		{name: "leading plus sign is stripped", input: "+52.34", want: 52.34},
		{name: "surrounding whitespace is trimmed", input: "  52.34  ", want: 52.34},
		{name: "negative with dollar sign", input: "-$52.34", want: -52.34},
		{name: "zero", input: "0", want: 0},
		{name: "integer without decimal", input: "100", want: 100},
		{name: "large comma separated negative", input: "-1,234,567.89", want: -1234567.89},
		{name: "invalid input returns error", input: "not-a-number", wantErr: true},
		{name: "empty string returns error", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmountFloat(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

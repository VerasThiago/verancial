package helper

import (
	"fmt"
	"strings"
)

func ParseAmountFloat(str string) (float32, error) {
	var value float32
	str = strings.Replace(str, "$", "", -1)
	str = strings.ReplaceAll(str, ",", "")
	str = strings.TrimPrefix(str, "+")
	str = strings.TrimSpace(str)

	if _, err := fmt.Sscanf(str, "%f", &value); err != nil {
		return 0, err
	}

	return value, nil
}

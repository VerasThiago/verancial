package helper

import (
	"fmt"
)

func ParseFloat(str string) (float32, error) {
	var value float32
	if _, err := fmt.Sscanf(str, "%f", &value); err != nil {
		return 0, err
	}
	return value, nil
}

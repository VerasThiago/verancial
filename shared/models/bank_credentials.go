package models

import (
	"encoding/json"
	"fmt"

	"github.com/verasthiago/verancial/shared/constants"
)

type BankCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type BankCredentialsMap map[constants.AppID]*BankCredentials

func (b *BankCredentialsMap) Scan(value interface{}) error {
	if value == nil {
		*b = nil
		return nil
	}
	byteSlice, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("BankCredentialsMap.Scan: unsupported value type %T", value)
	}
	var m map[string]map[string]interface{}
	err := json.Unmarshal(byteSlice, &m)
	if err != nil {
		return err
	}
	*b = make(BankCredentialsMap)
	for k, v := range m {
		fc := BankCredentials{
			Login:    v["login"].(string),
			Password: v["password"].(string),
		}
		(*b)[constants.AppID(k)] = &fc
	}
	return nil
}

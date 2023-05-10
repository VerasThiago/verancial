package models

import (
	"encoding/json"
	"fmt"

	"github.com/verasthiago/verancial/shared/constants"
)

type FinancialAppCredentials struct {
	Login    string                      `json:"login"`
	Password string                      `json:"password"`
	Metadata map[constants.BankId]string `json:"metadata" gorm:"type:jsonb"`
}

type FinancialAppCredentialsMap map[constants.AppID]*FinancialAppCredentials

func (f *FinancialAppCredentialsMap) Scan(value interface{}) error {
	if value == nil {
		*f = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("FinancialAppCredentialsMap.Scan: unsupported value type %T", value)
	}
	var m map[string]map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	*f = make(FinancialAppCredentialsMap)
	for k, v := range m {
		fc := FinancialAppCredentials{
			Login:    v["login"].(string),
			Password: v["password"].(string),
			Metadata: make(map[constants.BankId]string),
		}
		for metadataKey, metadataValue := range v["metadata"].(map[string]interface{}) {
			fc.Metadata[constants.BankId(metadataKey)] = metadataValue.(string)
		}
		(*f)[constants.AppID(k)] = &fc
	}
	return nil
}

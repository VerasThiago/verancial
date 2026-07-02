package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/verasthiago/verancial/shared/constants"
)

type FinancialAppCredentials struct {
	Login    string      `json:"login"`
	Password string      `json:"password"`
	Metadata interface{} `json:"metadata" gorm:"type:jsonb"`
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
	var m map[string]FinancialAppCredentials
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	*f = make(FinancialAppCredentialsMap)
	for k, v := range m {
		(*f)[constants.AppID(k)] = &v
	}
	return nil
}

// Value implements driver.Valuer so FinancialAppCredentialsMap round-trips
// through database/sql on its own (pairing the Scan above), rather than
// relying on the Postgres driver's jsonb-specific encoding fallback.
func (f FinancialAppCredentialsMap) Value() (driver.Value, error) {
	if f == nil {
		return nil, nil
	}
	return json.Marshal(f)
}

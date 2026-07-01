package models

import (
	"database/sql/driver"
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
		login, loginOk := v["login"].(string)
		password, passwordOk := v["password"].(string)
		if !loginOk || !passwordOk {
			return fmt.Errorf("BankCredentialsMap.Scan: entry %q missing string login/password", k)
		}
		(*b)[constants.AppID(k)] = &BankCredentials{Login: login, Password: password}
	}
	return nil
}

// Value implements driver.Valuer so BankCredentialsMap round-trips through
// database/sql on its own (pairing the Scan above), rather than relying on
// the Postgres driver's jsonb-specific encoding fallback.
func (b BankCredentialsMap) Value() (driver.Value, error) {
	if b == nil {
		return nil, nil
	}
	return json.Marshal(b)
}

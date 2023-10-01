package pkg

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// KeyValue is a key value map that can be stored as jsonb.
type KeyValue map[string]string

func (kv *KeyValue) Value() (driver.Value, error) {
	return json.Marshal(kv)
}

func (kv *KeyValue) Scan(value interface{}) error {
	var data []byte

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	case nil:
		return nil
	default:
		return errors.New("type assertion failed")
	}

	return json.Unmarshal(data, &kv)
}

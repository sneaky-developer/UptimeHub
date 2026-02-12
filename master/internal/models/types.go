package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// JSON is a custom type for JSONB columns
type JSON map[string]interface{}

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = JSON{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// UUIDArray is a custom type for UUID[] columns in PostgreSQL
type UUIDArray []uuid.UUID

func (a UUIDArray) Value() (driver.Value, error) {
	if a == nil || len(a) == 0 {
		return "{}", nil
	}
	result := "{"
	for i, id := range a {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%q", id.String())
	}
	result += "}"
	return result, nil
}

func (a *UUIDArray) Scan(value interface{}) error {
	if value == nil {
		*a = UUIDArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	str := string(bytes)
	if str == "{}" || str == "" {
		*a = UUIDArray{}
		return nil
	}

	// Parse PostgreSQL array format: {"uuid1","uuid2"}
	str = str[1 : len(str)-1] // remove { }
	var result UUIDArray
	for _, part := range splitPGArray(str) {
		id, err := uuid.Parse(part)
		if err != nil {
			return fmt.Errorf("failed to parse UUID %q: %w", part, err)
		}
		result = append(result, id)
	}
	*a = result
	return nil
}

func splitPGArray(s string) []string {
	var parts []string
	var current string
	inQuote := false
	for _, c := range s {
		switch {
		case c == '"':
			inQuote = !inQuote
		case c == ',' && !inQuote:
			parts = append(parts, current)
			current = ""
		default:
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

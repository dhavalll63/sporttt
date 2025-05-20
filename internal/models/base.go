// internal/models/base.go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type BaseModel struct {
	gorm.Model
}

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SocialMedia is the JSONB column for a user's social links.
type SocialMedia struct {
	Facebook  string `json:"facebook"`
	Instagram string `json:"instagram"`
	LinkedIn  string `json:"linkedin"`
	Twitter   string `json:"twitter"`
}

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan unmarshals a JSONB column into the struct.
func (s *StringSlice) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("SocialMedia: expected []byte, got %T", src)
	}
	return json.Unmarshal(b, s)
}
func (c Coordinates) Value() (driver.Value, error) {
	return json.Marshal(c)
}
func (c *Coordinates) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Coordinates: expected []byte, got %T", src)
	}
	return json.Unmarshal(b, c)
}
func (sm SocialMedia) Value() (driver.Value, error) {
	return json.Marshal(sm)
}

// Scan unmarshals JSONB bytes into the struct.
func (sm *SocialMedia) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("SocialMedia: expected []byte, got %T", src)
	}
	return json.Unmarshal(b, sm)
}

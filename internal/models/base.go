// internal/models/base.go
package models

import "gorm.io/gorm"

type BaseModel struct {
	gorm.Model
}

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type SocialMedia struct {
	Facebook  string `json:"facebook,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
}

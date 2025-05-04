package models

import "time"

type OptionBase struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Species OptionBase

type EggColor OptionBase

type EggSize OptionBase

type Coop OptionBase

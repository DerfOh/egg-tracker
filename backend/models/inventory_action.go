package models

import "time"

type InventoryAction struct {
	ID        int64     `json:"id"`
	Quantity  int       `json:"quantity"`
	Species   string    `json:"species"`
	Coop      string    `json:"coop"`
	EggColor  string    `json:"egg_color"`
	EggSize   string    `json:"egg_size"`
	Action    string    `json:"action"` // e.g., "collected", "sold", etc.
	Notes     *string   `json:"notes,omitempty"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

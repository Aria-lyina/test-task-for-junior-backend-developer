package task

import "time"

type Period struct {
    ID           int64     `json:"id"`
    Code         string    `json:"code"`
    Title        string    `json:"title"`
    RRULETemplate *string  `json:"rrule_template,omitempty"` // может быть NULL
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
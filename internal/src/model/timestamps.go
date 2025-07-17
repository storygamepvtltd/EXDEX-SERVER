package models

import (
	"time"
)

type Timestamps struct {
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type TimestampSetter interface {
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
}

func (t *Timestamps) SetCreatedAt(ts time.Time) {
	t.CreatedAt = ts
}

func (t *Timestamps) SetUpdatedAt(ts time.Time) {
	t.UpdatedAt = ts
}

package models

import (
	"time"
)

type User struct {
	ID            string     `bson:"_id,omitempty" json:"id"`
	ExdexUserID   uint       `bson:"exdex_user_id,omitempty" json:"exdex_user_id"`
	Email         string     `bson:"email" json:"email"`
	Password      string     `bson:"password,omitempty" json:"-"` // do not expose in JSON
	FullName      string     `bson:"full_name" json:"full_name"`
	AccountNumber string     `bson:"account_number" json:"account_number"`
	CreatedAt     time.Time  `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt     time.Time  `bson:"updated_at,omitempty" json:"updated_at"`
	DeletedAt     *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	Role          string     `bson:"role" json:"role"` // e.g., "user", "admin"
	SuspendedAt   *time.Time `bson:"suspended_at,omitempty" json:"suspended_at,omitempty"`
	Status        bool       `bson:"status" json:"status"` // e.g., "active", "inactive", "suspended"
}

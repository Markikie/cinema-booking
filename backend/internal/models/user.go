package models

import "time"

type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

type User struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	GoogleID  string    `bson:"google_id" json:"google_id"`
	Email     string    `bson:"email" json:"email"`
	Name      string    `bson:"name" json:"name"`
	Role      Role      `bson:"role" json:"role"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

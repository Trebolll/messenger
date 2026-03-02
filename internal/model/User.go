package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	FullName  string     `json:"full_name"`
	Password  string     `json:"password,omitempty"`
	BirthDate *time.Time `json:"birth_date"`
	Location  string     `json:"location"`
	Status    string     `json:"status"`
	AvatarUrl string     `json:"avatar_url"`
	Rating    int        `json:"rating"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

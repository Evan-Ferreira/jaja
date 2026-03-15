package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

var TestUser = User{
	ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}


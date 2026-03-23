package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrgID     *uuid.UUID `json:"org_id" gorm:"type:uuid"`
	Org       *Org       `json:"org,omitempty" gorm:"foreignKey:OrgID"`
	CreatedAt time.Time  `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"type:timestamptz;default:now()"`
}

func (User) TableName() string {
	return "users"
}


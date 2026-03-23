package models

import (
	"time"

	"github.com/google/uuid"
)

type Org struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrgName    string    `json:"org_name" gorm:"not null;uniqueIndex"`
	D2LOrgID   string    `json:"d2l_org_id" gorm:"not null;uniqueIndex"`
	D2LBaseURL string    `json:"d2l_base_url" gorm:"not null;uniqueIndex"`
	LEVersion  *string   `json:"le_version"`
	LPVersion  *string   `json:"lp_version"`
	CreatedAt  time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`
}

func (Org) TableName() string {
	return "orgs"
}

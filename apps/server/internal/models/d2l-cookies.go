package models

import (
	"time"

	"github.com/google/uuid"
)

type D2LCookies struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`
	UserId    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;references:users(id)"`
	Clck                string `json:"_clck" gorm:"column:clck;type:text"`
	Clsk                string `json:"_clsk" gorm:"column:clsk;type:text"`
	D2LSameSiteCanaryA  string `json:"d2lSameSiteCanaryA" gorm:"column:d2l_same_site_canary_a;type:text"`
	D2LSameSiteCanaryB  string `json:"d2lSameSiteCanaryB" gorm:"column:d2l_same_site_canary_b;type:text"`
	D2LSecureSessionVal string `json:"d2lSecureSessionVal" gorm:"column:d2l_secure_session_val;type:text"`
	D2LSessionVal       string `json:"d2lSessionVal" gorm:"column:d2l_session_val;type:text"`
}

func (D2LCookies) TableName() string {
	return "d2l_cookies"
}

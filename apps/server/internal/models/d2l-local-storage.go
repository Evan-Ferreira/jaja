package models

import (
	"time"

	"github.com/google/uuid"
)

type D2LLocalStorage struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`
	UserId    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;references:users(id)"`
	D2LFetchTokens      string `json:"D2L.Fetch.Tokens" gorm:"column:d2l_fetch_tokens;type:text"`
	SessionExpired      string `json:"Session.Expired" gorm:"column:session_expired;type:text"`
	SessionLastAccessed string `json:"Session.LastAccessed" gorm:"column:session_last_accessed;type:text"`
	SessionUserId       string `json:"Session.UserId" gorm:"column:session_user_id;type:text"`
	XsrfHitCodeSeed     string `json:"XSRF.HitCodeSeed" gorm:"column:xsrf_hit_code_seed;type:text"`
	XsrfToken           string `json:"XSRF.Token" gorm:"column:xsrf_token;type:text"`
	PdfjsHistory        string `json:"pdfjs.history" gorm:"column:pdfjs_history;type:text"`
}

func (D2LLocalStorage) TableName() string {
	return "d2l_local_storages"
}

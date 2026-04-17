package models

import (
	"encoding/json"
	"server/internal/jobs"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`
	Payload json.RawMessage `json:"payload,omitempty" gorm:"type:jsonb"`
	JobType jobs.JobType `json:"job_type" gorm:"type:text;not null"`
	Status  jobs.JobStatus  `json:"status" gorm:"type:text;not null"`
	Result  json.RawMessage `json:"result,omitempty" gorm:"type:jsonb"`
	Error   string `json:"error,omitempty" gorm:"type:text"`
}

func (Job) TableName() string {
	return "jobs"
}
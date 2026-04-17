package models

import (
	"time"
)

type Job struct {
	ID            string    `json:"id" gorm:"type:text;primaryKey"`
	Queue         string    `json:"queue" gorm:"type:text;not null;default:'default'"`
	Type          string    `json:"type" gorm:"type:text;not null"`
	Payload       []byte    `json:"payload,omitempty" gorm:"type:jsonb"`
	State         string    `json:"state" gorm:"type:text;not null"`
	MaxRetry      int       `json:"max_retry" gorm:"default:25"`
	Retried       int       `json:"retried" gorm:"default:0"`
	LastErr       string    `json:"last_err,omitempty" gorm:"type:text"`
	LastFailedAt  time.Time `json:"last_failed_at,omitempty" gorm:"type:timestamptz"`
	Result        []byte    `json:"result,omitempty" gorm:"type:bytea"`
	CompletedAt   time.Time `json:"completed_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt     time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`
}

func (Job) TableName() string {
	return "jobs"
}
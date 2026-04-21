package models

import (
	"encoding/json"
	"time"
)

type JobState string;

const (
	JobStatePending JobState = "pending"
	JobStateRunning JobState = "running"
	JobStateCompleted JobState = "completed"
	JobStateArchived JobState = "archived"
)

type Queue string;

const (
	QueueDefault Queue = "default"
)

type Job struct {
	ID            string    `json:"id" gorm:"type:text;primaryKey"`
	Queue         Queue     `json:"queue" gorm:"type:text;not null;default:'default'"`
	Type          string    `json:"type" gorm:"type:text;not null"`
	Payload       json.RawMessage    `json:"payload,omitempty" gorm:"type:jsonb"`
	State         JobState  `json:"state" gorm:"type:text;not null"`
	MaxRetry      int       `json:"max_retry" gorm:"default:25"`
	Retried       int       `json:"retried" gorm:"default:0"`
	LastErr       string    `json:"last_err,omitempty" gorm:"type:text"`
	LastFailedAt  time.Time `json:"last_failed_at,omitempty" gorm:"type:timestamptz"`
	Result        json.RawMessage    `json:"result,omitempty" gorm:"type:jsonb"`
	CompletedAt   time.Time `json:"completed_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt     time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`
}

func (Job) TableName() string {
	return "jobs"
}
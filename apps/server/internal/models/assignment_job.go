package models

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentJobStatus string

const (
	AssignmentJobStatusPending    AssignmentJobStatus = "pending"
	AssignmentJobStatusProcessing AssignmentJobStatus = "processing"
	AssignmentJobStatusCompleted  AssignmentJobStatus = "completed"
	AssignmentJobStatusFailed     AssignmentJobStatus = "failed"
	AssignmentJobStatusCancelled  AssignmentJobStatus = "cancelled"
	AssignmentJobStatusExpired    AssignmentJobStatus = "expired"
)

type AssignmentJob struct {
	ID               uuid.UUID           `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AssignmentID     uuid.UUID           `json:"assignment_id" gorm:"type:uuid;not null"`
	Assignment       *Assignment         `json:"assignment,omitempty" gorm:"foreignKey:AssignmentID"`
	UserID           uuid.UUID           `json:"user_id" gorm:"type:uuid;not null"`
	User             *User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	AnthropicBatchID *string             `json:"anthropic_batch_id" gorm:"index"`
	Status           AssignmentJobStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Result           *string             `json:"result"`
	ErrorMessage     *string             `json:"error_message"`
	CreatedAt        time.Time           `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt        time.Time           `json:"updated_at" gorm:"type:timestamptz;default:now()"`
}

func (AssignmentJob) TableName() string {
	return "assignment_jobs"
}

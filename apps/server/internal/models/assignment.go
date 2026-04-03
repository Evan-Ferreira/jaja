package models

import (
	"time"

	"github.com/google/uuid"
)

type Assignment struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CourseID         uuid.UUID  `json:"course_id" gorm:"type:uuid;uniqueIndex:idx_assignments_d2l_id_course_id;not null"`
	Course           *Course    `json:"course,omitempty" gorm:"foreignKey:CourseID"`
	D2LID            int        `json:"d2l_id" gorm:"column:d2l_id;uniqueIndex:idx_assignments_d2l_id_course_id;not null"`
	Name             string     `json:"name" gorm:"not null"`
	Description      *string    `json:"description"`
	InstructionsText *string    `json:"instructions_text"`
	DueDate          *time.Time `json:"due_date" gorm:"type:timestamptz"`
	ScoreOutOf       *float64   `json:"score_out_of"`
	IsHidden         bool       `json:"is_hidden" gorm:"default:false"`
	CreatedAt        time.Time  `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"type:timestamptz;default:now()"`
}

func (Assignment) TableName() string {
	return "assignments"
}

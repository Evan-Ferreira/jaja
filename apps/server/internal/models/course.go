package models

import (
	"time"

	"github.com/google/uuid"
)

type Course struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrgID         uuid.UUID  `json:"org_id" gorm:"type:uuid;not null"`
	Org           *Org       `json:"org,omitempty" gorm:"foreignKey:OrgID"`
	D2LID         int        `json:"d2l_id" gorm:"column:d2l_id;uniqueIndex;not null"`
	Name          string     `json:"name" gorm:"not null"`
	Code          string     `json:"code" gorm:"not null"`
	Description   *string    `json:"description"`
	SyllabusS3Key *string    `json:"syllabus_s3_key"` // object key in MinIO
	CreatedAt     time.Time  `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"type:timestamptz;default:now()"`
}

func (Course) TableName() string {
	return "courses"
}

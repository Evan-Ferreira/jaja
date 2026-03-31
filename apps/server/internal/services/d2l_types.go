package services

import (
	"fmt"
	"net/http"
)

// ── Errors ────────────────────────────────────────────────────────────────────

type d2lStatusError struct {
	code int
	path string
}

func (e *d2lStatusError) Error() string {
	return fmt.Sprintf("d2l: unexpected status %d for %s", e.code, e.path)
}

// ── D2L API response types ────────────────────────────────────────────────────

type d2lOrgUnitType struct {
	ID   int    `json:"Id"`
	Code string `json:"Code"`
}

type d2lOrgUnit struct {
	ID   int            `json:"Id"`
	Type d2lOrgUnitType `json:"Type"`
	Name string         `json:"Name"`
	Code string         `json:"Code"`
}

type d2lEnrollmentAccess struct {
	CanAccess bool `json:"CanAccess"`
}

type d2lEnrollment struct {
	OrgUnit d2lOrgUnit           `json:"OrgUnit"`
	Access  d2lEnrollmentAccess  `json:"Access"`
}

type d2lEnrollmentsPage struct {
	PagingInfo struct {
		Bookmark     string `json:"Bookmark"`
		HasMoreItems bool   `json:"HasMoreItems"`
	} `json:"PagingInfo"`
	Items []d2lEnrollment `json:"Items"`
}

type d2lRichText struct {
	Text string `json:"Text"`
	HTML string `json:"Html"`
}

type d2lDropboxAssessment struct {
	ScoreDenominator *float64 `json:"ScoreDenominator"`
}

type d2lAttachment struct {
	FileID   int    `json:"FileId"`
	FileName string `json:"FileName"`
}

type d2lDropboxFolder struct {
	ID           int                  `json:"Id"`
	Name         string               `json:"Name"`
	Instructions d2lRichText          `json:"CustomInstructions"`
	DueDate      *string              `json:"DueDate"` // ISO 8601 or null
	Assessment   d2lDropboxAssessment `json:"Assessment"`
	IsHidden     bool                 `json:"IsHidden"`
	Attachments  []d2lAttachment      `json:"Attachments"`
}

// ── Public output types ───────────────────────────────────────────────────────

type Assignment struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Instructions string   `json:"instructions"`
	DueDate      *string  `json:"due_date"`
	ScoreOutOf   *float64 `json:"score_out_of"` // sourced from Assessment.ScoreDenominator
}

type Course struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Code        string       `json:"code"`
	Assignments []Assignment `json:"assignments"`
}

type D2LClient struct {
	orgID     string
	leVersion string
	lpVersion string
	token     string
	baseURL   string
	http      *http.Client
}

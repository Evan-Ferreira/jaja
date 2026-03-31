package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"server/internal/config"
	"server/internal/models"

	"github.com/google/uuid"
)

// ── Constants ─────────────────────────────────────────────────────────────────

const (
	CourseOfferingOrgUnitTypeID = 3 // ID for course offerings in D2L's org unit type system

	EnrollmentsPath       = "/d2l/api/lp/%s/enrollments/myenrollments/?orgUnitTypeId=%d"
	DropboxFoldersPath    = "/d2l/api/le/%s/%d/dropbox/folders/"
	DropboxAttachmentPath = "/d2l/api/le/%s/%d/dropbox/folders/%d/attachments/%d"
)

// ── Client ────────────────────────────────────────────────────────────────────

func NewD2LClient(userID uuid.UUID) (*D2LClient, error) {
	var session models.D2LLocalStorageSession
	if result := config.DBClient.Where("user_id = ?", userID).Last(&session); result.Error != nil {
		return nil, fmt.Errorf("d2l: no session found for user: %w", result.Error)
	}

	if session.FetchAccessToken == "" {
		return nil, fmt.Errorf("d2l: access token is empty in stored session")
	}

	var user models.User
	if result := config.DBClient.Preload("Org").First(&user, "id = ?", userID); result.Error != nil {
		return nil, fmt.Errorf("d2l: no user found: %w", result.Error)
	}

	if user.Org == nil {
		return nil, fmt.Errorf("d2l: user has no associated org")
	}

	if user.Org.LEVersion == nil || *user.Org.LEVersion == "" {
		return nil, fmt.Errorf("d2l: LE API version not set for org")
	}

	if user.Org.LPVersion == nil || *user.Org.LPVersion == "" {
		return nil, fmt.Errorf("d2l: LP API version not set for org")
	}

	return &D2LClient{
		orgID:     user.Org.ID.String(),
		leVersion: *user.Org.LEVersion,
		lpVersion: *user.Org.LPVersion,
		token:     session.FetchAccessToken,
		baseURL:   user.Org.D2LBaseURL,
		http:      &http.Client{},
	}, nil
}

// getBytes fetches path and returns the raw response body.
func (c *D2LClient) getBytes(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("d2l: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("d2l: request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, &d2lStatusError{code: res.StatusCode, path: path}
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("d2l: read body: %w", err)
	}
	return body, nil
}

// get fetches path and JSON-decodes the response into out.
func (c *D2LClient) get(path string, out any) error {
	body, err := c.getBytes(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("d2l: decode response: %w", err)
	}
	return nil
}

// ── Methods ───────────────────────────────────────────────────────────────────

func isSkippableStatus(err error) bool {
	var e *d2lStatusError
	return errors.As(err, &e) && (e.code == http.StatusForbidden || e.code == http.StatusNotFound)
}

// getActiveEnrollments fetches all active course-offering enrollments for the authenticated user.
// D2L paginates results, so we loop until there are no more pages. Closed enrollments are filtered out.
func (c *D2LClient) getActiveEnrollments() ([]d2lEnrollment, error) {
	basePath := fmt.Sprintf(EnrollmentsPath, c.lpVersion, CourseOfferingOrgUnitTypeID)

	var all []d2lEnrollment
	var page d2lEnrollmentsPage

	if err := c.get(basePath, &page); err != nil {
		return nil, err
	}

	for _, e := range page.Items {
		if e.Access.CanAccess {
			all = append(all, e)
		}
	}

	for page.PagingInfo.HasMoreItems {
		if err := c.get(basePath+"&bookmark="+page.PagingInfo.Bookmark, &page); err != nil {
			return nil, err
		}
		for _, e := range page.Items {
			if e.Access.CanAccess {
				all = append(all, e)
			}
		}
	}

	return all, nil
}

// getAssignments fetches all visible dropbox folders for a given org unit and uploads
// any file attachments to MinIO S3 under assignments/{orgUnitID}/{folderID}/{fileName}.
// TO DO: Will need to exapand to other assignment types in the future, but dropbox folders are the only type with due dates, so we start here.
func (c *D2LClient) getAssignments(orgUnitID int) ([]d2lDropboxFolder, error) {
	path := fmt.Sprintf(DropboxFoldersPath, c.leVersion, orgUnitID)
	var folders []d2lDropboxFolder
	if err := c.get(path, &folders); err != nil {
		if isSkippableStatus(err) {
			return nil, nil
		}
		return nil, err
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	for _, f := range folders {
		for _, a := range f.Attachments {
			wg.Add(1)
			go func(folderID int, attachment d2lAttachment) {
				defer wg.Done()
				key := fmt.Sprintf("assignments/%d/%d/%s", orgUnitID, folderID, attachment.FileName)

				//TODO: Change test-buckert to real bucket name and handle bucket creation if it doesn't exist
				exists, err := config.S3BasicsBucket.ObjectExists(ctx, "test-bucket", key)
				if err != nil {
					log.Printf("d2l: check attachment %q in S3: %v", key, err)
					return
				}
				if exists {
					return
				}

				attachPath := fmt.Sprintf(DropboxAttachmentPath, c.leVersion, orgUnitID, folderID, attachment.FileID)
				data, err := c.getBytes(attachPath)
				if err != nil {
					log.Printf("d2l: download attachment %q (folder %d, org %d): %v", attachment.FileName, folderID, orgUnitID, err)
					return
				}

				if err := config.S3BasicsBucket.UploadLargeObject(ctx, "test-bucket", key, data); err != nil {
					log.Printf("d2l: upload attachment %q to S3: %v", key, err)
				}
			}(f.ID, a)
		}
	}
	wg.Wait()

	return folders, nil
}

// LoadCoursesAndAssignments returns all enrolled courses with their assignments,
// ready to send to the frontend. Assignment fetches run concurrently.
func (c *D2LClient) LoadCoursesAndAssignments() ([]Course, error) {
	enrollments, err := c.getActiveEnrollments()
	if err != nil {
		return nil, fmt.Errorf("d2l: load enrollments: %w", err)
	}

	courses := make([]Course, len(enrollments))
	errs := make(chan error, len(enrollments))

	var wg sync.WaitGroup
	for i, e := range enrollments {
		wg.Add(1)

		//SUII MULTITHREADING
		go func(idx int, enrollment d2lEnrollment) {
			defer wg.Done()
			folders, err := c.getAssignments(enrollment.OrgUnit.ID)
			if err != nil {
				errs <- fmt.Errorf("d2l: load assignments for org %d: %w", enrollment.OrgUnit.ID, err)
				return
			}

			assignments := make([]Assignment, 0, len(folders))
			for _, f := range folders {
				if f.IsHidden {
					continue
				}
				assignments = append(assignments, Assignment{
					ID:           f.ID,
					Name:         f.Name,
					Instructions: f.Instructions.Text,
					DueDate:      f.DueDate,
					ScoreOutOf:   f.Assessment.ScoreDenominator,
				})
			}

			courses[idx] = Course{
				ID:          enrollment.OrgUnit.ID,
				Name:        enrollment.OrgUnit.Name,
				Code:        enrollment.OrgUnit.Code,
				Assignments: assignments,
			}
		}(i, e)
	}
	wg.Wait()
	close(errs)

	if err := <-errs; err != nil {
		return nil, err
	}

	return courses, nil
}

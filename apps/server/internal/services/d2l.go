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
	"server/internal/util"

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

	//  D2L has no way of returning all enrollments at once its always paginated.
	// Each enrollment is visited exactly once across all pages.
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

// getAssignments fetches all dropbox folders for a given org unit.
// TO DO: Will need to expand to other assignment types in the future, but dropbox folders are the only type with due dates, so we start here.
func (c *D2LClient) getAssignments(orgUnitID int) ([]d2lDropboxFolder, error) {
	path := fmt.Sprintf(DropboxFoldersPath, c.leVersion, orgUnitID)
	var folders []d2lDropboxFolder
	if err := c.get(path, &folders); err != nil {
		if isSkippableStatus(err) {
			return nil, nil
		}
		return nil, err
	}
	return folders, nil
}

func (c *D2LClient) saveAttachment(ctx context.Context, orgUnitID int, folderID int, attachment d2lAttachment) {
	key := fmt.Sprintf("assignments/%d/%d/%s", orgUnitID, folderID, attachment.FileName)

	//TODO: Change test-bucket to real bucket name and handle bucket creation if it doesn't exist
	exists, err := config.S3BasicsBucket.ObjectExists(ctx, "test-bucket", key)
	if err != nil {
		log.Printf("d2l: check attachment %q in S3: %v", key, err)
		return
	}
	if exists {
		log.Printf("d2l: attachment with same filename %q already exists in S3, skipping upload", key)
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
}

// updateCourse upserts a single course into the DB and returns the persisted record.
func (c *D2LClient) updateCourse(orgUUID uuid.UUID, unit d2lOrgUnit) (models.Course, error) {
	var course models.Course
	result := config.DBClient.Where("d2l_id = ? AND org_id = ?", unit.ID, orgUUID).First(&course)
	if result.Error != nil {
		course = models.Course{
			OrgID: orgUUID,
			D2LID: unit.ID,
			Name:  unit.Name,
			Code:  unit.Code,
		}
		if err := config.DBClient.Create(&course).Error; err != nil {
			return course, fmt.Errorf("d2l: create course %d: %w", unit.ID, err)
		}
	} else {
		course.Name = unit.Name
		course.Code = unit.Code
		if err := config.DBClient.Save(&course).Error; err != nil {
			return course, fmt.Errorf("d2l: update course %d: %w", unit.ID, err)
		}
	}
	return course, nil
}

// updateAssignments fetches dropbox folders for orgUnitID from D2L and upserts them as assignments under course.
func (c *D2LClient) updateAssignments(course models.Course, orgUnitID int) error {
	folders, err := c.getAssignments(orgUnitID)
	if err != nil {
		return fmt.Errorf("d2l: load assignments for org %d: %w", orgUnitID, err)
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	for _, f := range folders {
		wg.Add(1)
		go func(folder d2lDropboxFolder) {
			defer wg.Done()
			for _, a := range folder.Attachments {
				c.saveAttachment(ctx, orgUnitID, folder.ID, a)
			}
			dueDate := util.ParseISODate(folder.DueDate)

			var assignment models.Assignment
			result := config.DBClient.Where("d2l_id = ? AND course_id = ?", folder.ID, course.ID).First(&assignment)
			if result.Error != nil {
				assignment = models.Assignment{
					CourseID:         course.ID,
					D2LID:            folder.ID,
					Name:             folder.Name,
					InstructionsText: &folder.Instructions.Text,
					DueDate:          dueDate,
					ScoreOutOf:       folder.Assessment.ScoreDenominator,
					IsHidden:         folder.IsHidden,
				}
				if err := config.DBClient.Create(&assignment).Error; err != nil {
					log.Printf("d2l: create assignment %d (course %d): %v", folder.ID, course.ID, err)
				}
			} else {
				assignment.Name = folder.Name
				assignment.InstructionsText = &folder.Instructions.Text
				assignment.DueDate = dueDate
				assignment.ScoreOutOf = folder.Assessment.ScoreDenominator
				assignment.IsHidden = folder.IsHidden
				if err := config.DBClient.Save(&assignment).Error; err != nil {
					log.Printf("d2l: update assignment %d (course %d): %v", folder.ID, course.ID, err)
				}
			}
		}(f)
	}
	wg.Wait()
	return nil
}

// SyncD2L fetches all enrolled courses and their assignments from D2L and upserts them into the DB.
// Course and assignment fetches run concurrently.
func (c *D2LClient) SyncD2L() error {
	enrollments, err := c.getActiveEnrollments()
	if err != nil {
		return fmt.Errorf("d2l: load enrollments: %w", err)
	}

	orgUUID, err := uuid.Parse(c.orgID)
	if err != nil {
		return fmt.Errorf("d2l: parse org ID: %w", err)
	}

	errs := make(chan error, len(enrollments))

	var wg sync.WaitGroup
	for _, e := range enrollments {
		wg.Add(1)

		//SUII MULTITHREADING
		go func(enrollment d2lEnrollment) {
			defer wg.Done()

			course, err := c.updateCourse(orgUUID, enrollment.OrgUnit)
			if err != nil {
				errs <- err
				return
			}

			if err := c.updateAssignments(course, enrollment.OrgUnit.ID); err != nil {
				errs <- err
			}
		}(e)
	}
	wg.Wait()
	close(errs)

	if err := <-errs; err != nil {
		return err
	}

	return nil
}

func (c *D2LClient) LoadCoursesAndAssignments() ([]Course, error) {
	orgUUID, err := uuid.Parse(c.orgID)
	if err != nil {
		return nil, fmt.Errorf("d2l: parse org ID: %w", err)
	}

	var dbCourses []models.Course
	if err := config.DBClient.Where("org_id = ?", orgUUID).Find(&dbCourses).Error; err != nil {
		return nil, fmt.Errorf("d2l: load courses from db: %w", err)
	}
	if len(dbCourses) == 0 {
		return []Course{}, nil
	}

	courseIDs := make([]uuid.UUID, len(dbCourses))
	for i, c := range dbCourses {
		courseIDs[i] = c.ID
	}

	var dbAssignments []models.Assignment
	if err := config.DBClient.Where("course_id IN ? AND is_hidden = false", courseIDs).Find(&dbAssignments).Error; err != nil {
		return nil, fmt.Errorf("d2l: load assignments from db: %w", err)
	}

	assignmentsByCourseID := make(map[uuid.UUID][]Assignment, len(dbCourses))
	for _, a := range dbAssignments {
		assignmentsByCourseID[a.CourseID] = append(assignmentsByCourseID[a.CourseID], Assignment{
			ID:           a.D2LID,
			Name:         a.Name,
			Instructions: a.InstructionsText,
			DueDate:      a.DueDate,
			ScoreOutOf:   a.ScoreOutOf,
		})
	}

	courses := make([]Course, len(dbCourses))
	for i, dbCourse := range dbCourses {
		assignments := assignmentsByCourseID[dbCourse.ID]
		if assignments == nil {
			assignments = []Assignment{}
		}
		courses[i] = Course{
			ID:          dbCourse.D2LID,
			Name:        dbCourse.Name,
			Code:        dbCourse.Code,
			Assignments: assignments,
		}
	}

	return courses, nil
}

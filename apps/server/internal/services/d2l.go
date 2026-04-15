package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"server/internal/storage"
	"sync"

	"server/internal/database"
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
	ContentTOCPath        = "/d2l/api/le/%s/%d/content/toc"
	ContentTopicFilePath  = "/d2l/api/le/%s/%d/content/topics/%d/file"
)

// ── Client ────────────────────────────────────────────────────────────────────

func NewD2LClient(userID uuid.UUID) (*D2LClient, error) {
	var session models.D2LLocalStorageSession
	if result := database.DBClient.Where("user_id = ?", userID).Last(&session); result.Error != nil {
		return nil, fmt.Errorf("d2l: no session found for user: %w", result.Error)
	}

	if session.FetchAccessToken == "" {
		return nil, fmt.Errorf("d2l: access token is empty in stored session")
	}

	var user models.User
	if result := database.DBClient.Preload("Org").First(&user, "id = ?", userID); result.Error != nil {
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

// get fetches path and returns the raw body and filename from Content-Disposition (empty if absent).
// If out is non-nil, the body is also JSON-decoded into it.
func (c *D2LClient) get(path string, out any) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, "", fmt.Errorf("d2l: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("d2l: request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, "", &d2lStatusError{code: res.StatusCode, path: path}
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "", fmt.Errorf("d2l: read body: %w", err)
	}

	var filename string
	if cd := res.Header.Get("Content-Disposition"); cd != "" {
		if _, params, parseErr := mime.ParseMediaType(cd); parseErr == nil {
			filename = params["filename"]
		}
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return nil, "", fmt.Errorf("d2l: decode response: %w", err)
		}
	}

	return body, filename, nil
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

	if _, _, err := c.get(basePath, &page); err != nil {
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
		if _, _, err := c.get(basePath+"&bookmark="+page.PagingInfo.Bookmark, &page); err != nil {
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
	if _, _, err := c.get(path, &folders); err != nil {
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
	exists, err := storage.S3BasicsBucket.ObjectExists(ctx, "test-bucket", key)
	if err != nil {
		log.Printf("d2l: check attachment %q in S3: %v", key, err)
		return
	}
	if exists {
		log.Printf("d2l: attachment with same filename %q already exists in S3, skipping upload", key)
		return
	}

	attachPath := fmt.Sprintf(DropboxAttachmentPath, c.leVersion, orgUnitID, folderID, attachment.FileID)
	data, _, err := c.get(attachPath, nil)
	if err != nil {
		log.Printf("d2l: download attachment %q (folder %d, org %d): %v", attachment.FileName, folderID, orgUnitID, err)
		return
	}

	if err := storage.S3BasicsBucket.UploadLargeObject(ctx, "test-bucket", key, data); err != nil {
		log.Printf("d2l: upload attachment %q to S3: %v", key, err)
	}
}

// collectTopics recursively flattens all topics from a TOC tree.
func collectTopics(modules []d2lContentModule, topics []d2lContentTopic) []d2lContentTopic {
	all := append([]d2lContentTopic{}, topics...)
	for _, m := range modules {
		all = append(all, collectTopics(m.Modules, m.Topics)...)
	}
	return all
}

// UpdateContent fetches the full content TOC for orgUnitID, then concurrently downloads
// every topic file and uploads it to S3 under content/{orgUnitID}/{topicID}.
// Non-accessible topics (403/404) are silently skipped.
func (c *D2LClient) UpdateContent(ctx context.Context, orgUnitID int) {
	tocPath := fmt.Sprintf(ContentTOCPath, c.leVersion, orgUnitID)

	var toc d2lContentTOC
	if _, _, err := c.get(tocPath, &toc); err != nil {
		if !isSkippableStatus(err) {
			log.Printf("d2l: fetch content TOC for org unit %d: %v", orgUnitID, err)
		}
		return
	}

	topics := collectTopics(toc.Modules, toc.Topics)
	log.Printf("d2l: found %d topics in TOC for org unit %d", len(topics), orgUnitID)

	var wg sync.WaitGroup
	for _, topic := range topics {
		wg.Add(1)
		go func(t d2lContentTopic) {
			defer wg.Done()

			filePath := fmt.Sprintf(ContentTopicFilePath, c.leVersion, orgUnitID, t.TopicID)
			data, filename, err := c.get(filePath, nil)
			if err != nil {
			if isSkippableStatus(err) {
					log.Printf("d2l: skipping topic %d %q (org unit %d): %v", t.TopicID, t.Title, orgUnitID, err)
				} else {
					log.Printf("d2l: download topic %d %q (org unit %d): %v", t.TopicID, t.Title, orgUnitID, err)
				}
				return
			}

			ext := filepath.Ext(filename)
			key := fmt.Sprintf("content/%d/%d_%s%s", orgUnitID, t.TopicID, storage.SanitizeS3Key(t.Title), ext)
			//TODO: Change test-bucket to real bucket name
			if err := storage.S3BasicsBucket.UploadLargeObject(ctx, "test-bucket", key, data); err != nil {
				log.Printf("d2l: upload topic %d %q to S3: %v", t.TopicID, t.Title, err)
				return
			}
			log.Printf("d2l: uploaded topic %d %q (%d bytes) -> %q", t.TopicID, t.Title, len(data), key)
		}(topic)
	}
	wg.Wait()
}

// updateCourse upserts a single course into the DB and returns the persisted record.
func (c *D2LClient) updateCourse(orgUUID uuid.UUID, unit d2lOrgUnit) (models.Course, error) {
	var course models.Course
	result := database.DBClient.Where("d2l_id = ? AND org_id = ?", unit.ID, orgUUID).First(&course)
	if result.Error != nil {
		course = models.Course{
			OrgID: orgUUID,
			D2LID: unit.ID,
			Name:  unit.Name,
			Code:  unit.Code,
		}
		if err := database.DBClient.Create(&course).Error; err != nil {
			return course, fmt.Errorf("d2l: create course %d: %w", unit.ID, err)
		}
	} else {
		course.Name = unit.Name
		course.Code = unit.Code
		if err := database.DBClient.Save(&course).Error; err != nil {
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
			result := database.DBClient.Where("d2l_id = ? AND course_id = ?", folder.ID, course.ID).First(&assignment)
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
				if err := database.DBClient.Create(&assignment).Error; err != nil {
					log.Printf("d2l: create assignment %d (course %d): %v", folder.ID, course.ID, err)
				}
			} else {
				assignment.Name = folder.Name
				assignment.InstructionsText = &folder.Instructions.Text
				assignment.DueDate = dueDate
				assignment.ScoreOutOf = folder.Assessment.ScoreDenominator
				assignment.IsHidden = folder.IsHidden
				if err := database.DBClient.Save(&assignment).Error; err != nil {
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

			//TODO: Add once we have job qeuee set up to avoid long-running request timeouts
			// ctx := context.Background()
			// c.UpdateContent(ctx, enrollment.OrgUnit.ID)

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
	if err := database.DBClient.Where("org_id = ?", orgUUID).Find(&dbCourses).Error; err != nil {
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
	if err := database.DBClient.Where("course_id IN ? AND is_hidden = false", courseIDs).Find(&dbAssignments).Error; err != nil {
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

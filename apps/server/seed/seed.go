package seed

import (
	"log"
	"os"
	"time"

	"server/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TestUserID is the hardcoded user ID used until real auth is implemented.
var TestUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
var TestOrgID = uuid.MustParse("00000000-0000-0000-0000-000000000010")

var TestOrg = models.Org{
	ID:         TestOrgID,
	OrgName:    "Test Organization",
	D2LBaseURL: "https://test.desire2learn.com",
	LEVersion:  "1.0",
	LPVersion:  "1.0",
	CreatedAt:  time.Now(),
	UpdatedAt:  time.Now(),
}

var TestUser = models.User{
	ID:        TestUserID,
	OrgID:     &TestOrgID,
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

func buildTestLocalStorage() models.D2LLocalStorageSession {
	return models.D2LLocalStorageSession{
		ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		UserId:              TestUserID,
		FetchAccessToken:    os.Getenv("D2L_TEST_ACCESS_TOKEN"),
		FetchExpiresAt:      9999999999,
		SessionExpired:      "false",
		SessionLastAccessed: "1700000000000",
		SessionUserId:       "12345",
		XsrfHitCodeSeed:     "test-seed",
		XsrfToken:           "test-xsrf-token",
		PdfjsHistory:        "{}",
	}
}

var TestCookie = models.D2LCookieSession{
	ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000003"),
	UserId:              TestUserID,
	Clck:                "test-clck",
	Clsk:                "test-clsk",
	D2LSameSiteCanaryA:  "test-canary-a",
	D2LSameSiteCanaryB:  "test-canary-b",
	D2LSecureSessionVal: "test-secure-session",
	D2LSessionVal:       "test-session",
}

func insert(db *gorm.DB, label string, value any) {
	result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(value)
	if result.Error != nil {
		log.Fatalf("seed: %s: %v", label, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Printf("seed: %s: already exists, skipped", label)
	} else {
		log.Printf("seed: %s: inserted (%d row)", label, result.RowsAffected)
	}
}

func Run(db *gorm.DB) {
	localStorage := buildTestLocalStorage()
	insert(db, "test org", &TestOrg)
	insert(db, "test user", &TestUser)
	insert(db, "test local storage", &localStorage)
	insert(db, "test cookie", &TestCookie)
	log.Println("seed: done")
}

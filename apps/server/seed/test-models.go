package seed

import (
	"time"

	"server/internal/models"

	"github.com/google/uuid"
)

const testFetchTokens = `{"*:*:*":{"expires_at":1773890484,"access_token":"eyJhbGciOiJSUzI1NiIsImtpZCI6IjM0OGIxMTY0LTkzY2ItNGM3My1iNmJkLWE1Mzg0YjAyYjQ1YSIsInR5cCI6IkpXVCJ9.eyJuYmYiOjE3NzM4ODY4ODQsImV4cCI6MTc3Mzg5MDQ4NCwiaXNzIjoiaHR0cHM6Ly9hcGkuYnJpZ2h0c3BhY2UuY29tL2F1dGgiLCJhdWQiOiJodHRwczovL2FwaS5icmlnaHRzcGFjZS5jb20vYXV0aC90b2tlbiIsInRlbmFudGlkIjoiM2Y1MDY5YjgtMDQ2NS00OTA5LTgxNGQtNGM5YzdiNzg1ZjZkIiwic3ViIjoiMzkyODYxIiwiYXpwIjoiZDJsLWlhbS1sbXMiLCJzY29wZSI6Iio6KjoqIiwianRpIjoiMWFjN2FjNDYtYjY2NS00YTc3LWE2NDctMjQ3YTE2MjUyZDQ2In0.dQlegDyW0RoaD8Obtdw4nRh37arKsl2LM0gMMrFpIs9a_EBFEW7wl2ZbkOwcEMEg7EjZ67SCn6UFYn_dqgNRC9AZOo57ypac-wRf8FkvXGSIH0AQoCK6Anc1pdX4gtdHgxckMZXG00Z3Kidbi15petqFbkzoql3ttVFC0faOXR4ictz2f1i4FTxr24ohSXs9CMuNepirVFUkU1JWnn0ZS-TQWZ2_fmMtlTiFLEyze1EZsTA-B1ulG1ugmbCI8Ncby5UPDrFYvkPnoOg0BodsxL10WIHBINvZWRfiP0nBt3WujH5eB0Iat7oikMpuNNBG7Nwr586-9prKDYwfDgVxRg"}}`

// TestUserID is the hardcoded user ID used until real auth is implemented.
var TestUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

var TestUser = models.User{
	ID:        TestUserID,
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

var TestLocalStorage = models.D2LLocalStorageSession{
	ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000002"),
	UserId:              TestUserID,
	D2LFetchTokens:      testFetchTokens,
	SessionExpired:      "false",
	SessionLastAccessed: "1700000000000",
	SessionUserId:       "12345",
	XsrfHitCodeSeed:     "test-seed",
	XsrfToken:           "test-xsrf-token",
	PdfjsHistory:        "{}",
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

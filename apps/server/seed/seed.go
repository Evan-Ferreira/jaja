package seed

import (
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
	insert(db, "test user", &TestUser)
	insert(db, "test local storage", &TestLocalStorage)
	insert(db, "test cookie", &TestCookie)
	log.Println("seed: done")
}

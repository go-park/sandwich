package lib

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func _init() {
	db, _ = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
}

func GetGormDB() *gorm.DB {
	return db
}

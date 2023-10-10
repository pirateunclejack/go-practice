package database

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/jinzhu/gorm"
)

var (
	DBConn *gorm.DB
	
)
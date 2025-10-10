package mysql

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

var gDSN string
var gModelPtrs []any
var gDB *gorm.DB
var mu sync.Mutex

func Init(dsn string, modelPtrs ...any) {
	gDSN = dsn
	gModelPtrs = modelPtrs
}

func Connect() (*gorm.DB, error) {
	mu.Lock()
	defer mu.Unlock()

	if gDB != nil {
		return gDB, nil
	}
	var err error
	gDB, err = doConnect()
	return gDB, err
}

func doConnect() (*gorm.DB, error) {
	if gDSN == "" {
		return nil, fmt.Errorf("dsn is empty")
	}
	db, err := gorm.Open(mysql.Open(gDSN), &gorm.Config{
		Logger: &Logger{},
	})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(gModelPtrs...)
	if err != nil {
		return nil, err
	}
	return db, nil
}

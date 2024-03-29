package db

import (
	"reflect"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type S3 struct {
	gdb    *gorm.DB
	locker sync.RWMutex
}

func NewS3(url string, debug bool) (*S3, error) {
	gConf := &gorm.Config{}
	if !debug {
		gConf.Logger = logger.Default.LogMode(logger.Silent)
	}
	db, err := gorm.Open(sqlite.Open(url), gConf)
	if err != nil {
		return nil, err
	}

	return &S3{gdb: db}, nil
}

func (db *S3) RawDb() *gorm.DB {
	return db.gdb
}

func (db *S3) Insert(items interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Create(items).Error
}

func (db *S3) InsertBatch(items ...interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	tx := db.gdb.Begin()
	for _, item := range items {
		if err := tx.CreateInBatches(item, reflect.ValueOf(item).Len()).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

// db.Where("uuid = ?", "test").Delete(&xxx)
func (db *S3) Delete(query interface{}, args interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Unscoped().Where(query, args).Delete(m).Error
}

func (db *S3) Update(query interface{}, args interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Model(m).Where(query, args).Updates(m).Error
}

func (db *S3) UpdateRaw(sql string, value interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()
	return db.gdb.Model(m).Raw(sql, value).Where("1=?", 1).Updates(m).Error
}

func (db *S3) First(query interface{}, args interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	tx := db.gdb.Where(query, args).First(m)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *S3) Find(query interface{}, args interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	tx := db.gdb.Where(query, args).Find(m)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *S3) FindRaw(sql string, m interface{}, values ...interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	tx := db.gdb.Raw(sql, values).Find(m)
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

package db

import (
	"reflect"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type S3DB struct {
	gdb    *gorm.DB
	locker sync.RWMutex
}

func NewS3DB(url string, debug bool) (*S3DB, error) {
	gConf := &gorm.Config{}
	if !debug {
		gConf.Logger = logger.Default.LogMode(logger.Silent)
	}
	db, err := gorm.Open(sqlite.Open(url), gConf)
	if err != nil {
		return nil, err
	}

	return &S3DB{gdb: db}, nil
}

func (db *S3DB) RawDb() *gorm.DB {
	return db.gdb
}

func (db *S3DB) Insert(items interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Create(items).Error
}

func (db *S3DB) InsertBatch(items ...interface{}) error {
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
func (db *S3DB) Remove(query interface{}, args interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Unscoped().Where(query, args).Delete(m).Error
}

func (db *S3DB) Update(query interface{}, args interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Model(m).Where(query, args).Updates(m).Error
}

func (db *S3DB) Update2(q1 interface{}, args1 interface{},
	q2 interface{}, args2 interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Model(m).Where(q1, args1).Where(q2, args2).Updates(m).Error
}

func (db *S3DB) First(query interface{}, args interface{}, m interface{}) *gorm.DB {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Where(query, args).First(m)
}

func (db *S3DB) Find(query interface{}, args interface{}, m interface{}) *gorm.DB {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Where(query, args).Find(m)
}

func (db *S3DB) Find2(q1 interface{}, a1 interface{}, q2 interface{}, a2 interface{}, m interface{}) *gorm.DB {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Where(q1, a1).Where(q2, a2).Find(m)
}

func (db *S3DB) FindWithOrder(query, args, order, m interface{}) *gorm.DB {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Where(query, args).Order(order).Find(m)
}

func (db *S3DB) FindRaw(sql string, limit int, m interface{}, values ...interface{}) *gorm.DB {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Raw(sql, values).Limit(limit).Find(m)
}

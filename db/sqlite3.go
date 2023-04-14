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
func (db *S3) Remove(query interface{}, args interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Unscoped().Where(query, args).Delete(m).Error
}

func (db *S3) Update(query interface{}, args interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Model(m).Where(query, args).Updates(m).Error
}

func (db *S3) Update2(q1 interface{}, args1 interface{},
	q2 interface{}, args2 interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Model(m).Where(q1, args1).Where(q2, args2).Updates(m).Error
}

func (db *S3) First(query interface{}, args interface{}, m interface{}) *gorm.DB {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Where(query, args).First(m)
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

func (db *S3) Find2(q1 interface{}, a1 interface{}, q2 interface{}, a2 interface{}, m interface{}) error {
	db.locker.Lock()
	defer db.locker.Unlock()

	tx := db.gdb.Where(q1, a1).Where(q2, a2).Find(m)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *S3) FindWithOrder(query, args, order, m interface{}) *gorm.DB {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Where(query, args).Order(order).Find(m)
}

func (db *S3) FindRaw(sql string, limit int, m interface{}, values ...interface{}) *gorm.DB {
	db.locker.Lock()
	defer db.locker.Unlock()

	return db.gdb.Raw(sql, values).Find(m).Limit(limit)
}

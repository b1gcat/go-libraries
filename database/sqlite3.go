package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
)

type Sqlite3 struct {
	once   sync.Once
	db     *gorm.DB
	locker sync.RWMutex
}

func NewSqlite3(url string, debug bool) (Database, error) {
	gConf := &gorm.Config{}
	if !debug {
		gConf.Logger = logger.Default.LogMode(logger.Silent)
	}
	db, err := gorm.Open(sqlite.Open(url), gConf)
	if err != nil {
		return nil, err
	}

	return &Sqlite3{db: db}, nil
}

func (sl *Sqlite3) DB() *gorm.DB {
	return sl.db
}

func (sl *Sqlite3) Insert(items interface{}) error {
	sl.locker.Lock()
	defer sl.locker.Unlock()

	return sl.db.Create(items).Error
}

func (sl *Sqlite3) Remove(condition string, items interface{}) error {
	sl.locker.Lock()
	defer sl.locker.Unlock()

	return sl.db.Where(condition).Delete(items).Error
}

func (sl *Sqlite3) Update(condition string, items interface{}, ignoreItems ...string) error {
	sl.locker.Lock()
	defer sl.locker.Unlock()

	m := remoteIgnore(items, ignoreItems...)
	return sl.db.Model(items).Where(condition).Updates(m).Error
}

func (sl *Sqlite3) Search(rawQuery string, items interface{}) (int64, error) {
	sl.locker.Lock()
	defer sl.locker.Unlock()

	r := sl.db.Raw(rawQuery).Scan(items)
	return r.RowsAffected, r.Error
}

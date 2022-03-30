package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
)

type Mysql struct {
	once sync.Once
	db   *gorm.DB
}

//NewMysql @url:kickYouAbc:kickYouAbc@tcp(106.75.174.5:2345)/unknown?charset=utf8&parseTime=True&loc=Local&&readTimeout=5s
func NewMysql(url string, debug bool) (Database, error) {
	gConf := &gorm.Config{}
	if !debug {
		gConf.Logger = logger.Default.LogMode(logger.Silent)
	}
	db, err := gorm.Open(mysql.Open(url), gConf)
	if err != nil {
		return nil, err
	}

	return &Mysql{db: db}, nil
}

func (sl *Mysql) DB() *gorm.DB {
	return sl.db
}

func (sl *Mysql) Insert(items interface{}) error {
	return sl.db.Create(items).Error
}

func (sl *Mysql) Remove(condition string, items interface{}) error {
	return sl.db.Where(condition).Delete(items).Error
}

func (sl *Mysql) Update(condition string, items interface{}, ignoreItems ...string) error {
	m := remoteIgnore(items, ignoreItems...)
	return sl.db.Model(items).Where(condition).Updates(m).Error
}

func (sl *Mysql) Search(rawQuery string, items interface{}) (int64, error) {
	r := sl.db.Raw(rawQuery).Scan(items)
	return r.RowsAffected, r.Error
}

package db

import (
	"reflect"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Mysql struct {
	gdb *gorm.DB
}

// NewMysql @url:kickYouAbc:kickYouAbc@tcp(106.75.174.5:2345)/unknown?charset=utf8&parseTime=True&loc=Local&&readTimeout=5s
func NewMysql(url string, debug bool) (*Mysql, error) {
	gConf := &gorm.Config{}
	if !debug {
		gConf.Logger = logger.Default.LogMode(logger.Silent)
	}
	db, err := gorm.Open(mysql.Open(url), gConf)
	if err != nil {
		return nil, err
	}

	return &Mysql{gdb: db}, nil
}

func (db *Mysql) RawDb() *gorm.DB {
	return db.gdb
}

func (db *Mysql) Insert(items interface{}) error {
	return db.gdb.Create(items).Error
}

func (db *Mysql) InsertBatch(items ...interface{}) error {
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
func (db *Mysql) Delete(query interface{}, args interface{}, m interface{}) error {
	return db.gdb.Unscoped().Where(query, args).Delete(m).Error
}

func (db *Mysql) Update(query interface{}, args interface{}, m interface{}) error {
	return db.gdb.Model(m).Where(query, args).Updates(m).Error
}

func (db *Mysql) Update2(q1 interface{}, args1 interface{},
	q2 interface{}, args2 interface{}, m interface{}) error {
	return db.gdb.Model(m).Where(q1, args1).Where(q2, args2).Updates(m).Error
}

func (db *Mysql) First(query interface{}, args interface{}, m interface{}) *gorm.DB {
	return db.gdb.Where(query, args).First(m)
}

func (db *Mysql) Find(query interface{}, args interface{}, m interface{}) error {
	tx := db.gdb.Where(query, args).Find(m)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *Mysql) FindWithOrder(query, args, order, m interface{}) *gorm.DB {
	return db.gdb.Where(query, args).Order(order).Find(m)
}

func (db *Mysql) FindRaw(sql string, m interface{}, values ...interface{}) *gorm.DB {
	return db.gdb.Raw(sql, values).Find(m)
}

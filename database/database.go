package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Database interface {
	DB() *gorm.DB
	Insert(items interface{}) error
	Remove(condition string, items interface{}) error
	Update(condition string, items interface{}, ignoreItems ...string) error
	Search(rawQuery string, items interface{}) (int64, error)
}

const (
	QName  = "qName"
	QType  = "qType"
	QValue = "qValue"
	QOrder = "qOrder"
	QLimit = "qLimit"
	QRaw   = "qRaw"
)

//MakeQueryCondition say...
//qName,qType,qValue,qOrder,qLimit
func MakeQueryCondition(q map[string]string) (condition string) {
	var qName, qType, qValue, qOrder, qLimit string
	for k, v := range q {
		switch k {
		case QName:
			qName = v
		case QType:
			qType = v
		case QValue:
			qValue = v
		case QOrder:
			qOrder = v
		case QLimit:
			qLimit = v
		case QRaw:
			return v
		}
	}
	switch qType {
	case "like":
		condition = fmt.Sprintf(`%s like '%%%s%%'`, qName, qValue)
	default:
		condition = fmt.Sprintf(`%s %s '%s'`, qName, qType, qValue)
	}

	if qOrder != "" {
		condition += " order by " + qOrder
	}
	if qLimit != "" {
		condition += " limit " + qLimit
	}
	return
}

func remoteIgnore(items interface{}, ignoreItems ...string) map[string]interface{} {
	aJson, _ := json.Marshal(items)
	var m map[string]interface{}
	_ = json.Unmarshal(aJson, &m)

	//Note: unique要去掉，不然更新失败
	for _, ignore := range ignoreItems {
		delete(m, ignore)
	}

	return m
}

type MyTime struct {
	time.Time
}

func (t MyTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
	return []byte(formatted), nil
}

func (t MyTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

func (t *MyTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = MyTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

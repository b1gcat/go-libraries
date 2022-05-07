package iot

import (
	"fmt"
)

type Null struct {
	Options
}

func (a *Null) Config() map[string]interface{} {
	return map[string]interface{}{
		"region": "null",
		"host":   "null",
		"ca":     []string{"null"},
	}
}

func NewNull(optFn ...OptionsFunc) (Iot, error) {
	options := defaultOption()
	for _, o := range optFn {
		if err := o(&options); err != nil {
			return nil, err
		}
	}
	return Iot(&Null{
		Options: options,
	}), nil
}

func (a *Null) Close() {
}

func (a *Null) Subscribe(s map[string]interface{}) error {
	return fmt.Errorf("no support")
}

//Publish say...
//Input：
//@message
//@topic
func (a *Null) Publish(s map[string]interface{}) error {
	a.logger.Debug("Publish:", s["message"].(string))
	return nil
}

//CreateThing 控制台生成物料以及身份认证信息
//Input：
//@tName
//@tType
//@tGroup
//@cert_arn
//Output:
//@arn
//@ak...
func (a *Null) CreateThing(t map[string]interface{}) (map[string]interface{}, error) {
	ret := map[string]interface{}{
		"arn":     "null",
		"pKey":    "null",
		"cert":    "null", //accessKeyId
		"cert_id": "null",
	}

	a.logger.Debug(fmt.Sprintf("AddThing:%v", t["tName"].(string)))
	return ret, nil
}

package iot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	iot20180120 "github.com/alibabacloud-go/iot-20180120/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"io/ioutil"
	"os"
	"path/filepath"
)

type aliConf struct {
	ProductKey   string `json:"product_key"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
}

type Ali struct {
	Options
	aliConf

	thing      *iot20180120.Client
	instanceId string
	region     string
}

func (a *Ali) Config() map[string]interface{} {
	return map[string]interface{}{
		"region": a.region,
		//xxx.iot-as-mqtt.yyy.aliyuncs.com
		"host": fmt.Sprintf("%s.iot-as-mqtt.%s.aliyuncs.com", a.ProductKey, a.region),
		"ca":   "",
	}
}

func NewAli(opt map[string]interface{}, optFn ...OptionsFunc) (Iot, error) {
	options := defaultOption()
	for _, o := range optFn {
		if err := o(&options); err != nil {
			return nil, err
		}
	}
	a := &Ali{
		Options: options,
		region:  opt["endpoint"].(string),
	}

	if err := a.aliConf.load(opt["profile"].(string)); err != nil {
		return nil, err
	}

	config := &openapi.Config{
		// 您的AccessKey ID
		AccessKeyId: tea.String(a.AccessKey),
		// 您的AccessKey Secret
		AccessKeySecret: tea.String(a.AccessSecret),
	}
	// 访问的域名
	config.Endpoint = tea.String(fmt.Sprintf("iot.%s.aliyuncs.com", a.region))

	var err error
	if a.thing, err = iot20180120.NewClient(config); err != nil {
		return nil, err
	}

	return Iot(a), nil
}

func (a *Ali) Close() {
	a.cancel()
}

//Subscribe say...
//Input：
//@endpoint
//@protocol
//@topic_arn
func (a *Ali) Subscribe(s map[string]interface{}) error {
	return fmt.Errorf("no support")
}

//Publish say...
//Input：
//@message
//@topic
func (a *Ali) Publish(s map[string]interface{}) error {
	pubRequest := &iot20180120.PubRequest{
		TopicFullName:  tea.String(fmt.Sprintf("/%s/%s/user/command", a.ProductKey, s["topic"].(string))),
		ProductKey:     tea.String(a.ProductKey),
		Qos:            tea.Int32(0),
		MessageContent: tea.String(base64.StdEncoding.EncodeToString([]byte(s["message"].(string)))),
	}

	// 复制代码运行请自行打印 API 的返回值
	if _, err := a.thing.Pub(pubRequest); err != nil {
		return err
	}
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
func (a *Ali) CreateThing(t map[string]interface{}) (map[string]interface{}, error) {
	//查询组信息
	tag := &iot20180120.QueryDeviceGroupByTagsRequestTag{
		TagValue: tea.String(t["tGroup"].(string)),
		TagKey:   tea.String(t["tGroup"].(string)),
	}
	queryDeviceGroupByTagsRequest := &iot20180120.QueryDeviceGroupByTagsRequest{
		Tag: []*iot20180120.QueryDeviceGroupByTagsRequestTag{tag},
	}
	grpInfo, err := a.thing.QueryDeviceGroupByTags(queryDeviceGroupByTagsRequest)
	if err != nil {
		return nil, err
	}
	//删除旧设备
	deleteDeviceRequest := &iot20180120.DeleteDeviceRequest{
		ProductKey: tea.String(a.ProductKey),
		DeviceName: tea.String(t["tName"].(string)),
	}
	_, _ = a.thing.DeleteDevice(deleteDeviceRequest)
	//创建新设备
	registerDeviceRequest := &iot20180120.RegisterDeviceRequest{
		DeviceName: tea.String(t["tName"].(string)),
		ProductKey: tea.String(a.ProductKey),
		Nickname:   tea.String(t["tType"].(string)),
	}

	devInfo, err := a.thing.RegisterDevice(registerDeviceRequest)
	if err != nil {
		return nil, err
	}

	a.logger.Debug(fmt.Sprintf("CreateThing:%s", *registerDeviceRequest.DeviceName))

	//加入组
	batchAddDeviceGroupRelationsRequest := &iot20180120.BatchAddDeviceGroupRelationsRequest{
		GroupId: grpInfo.Body.Data.DeviceGroup[0].GroupId,
		Device: []*iot20180120.BatchAddDeviceGroupRelationsRequestDevice{
			{
				DeviceName: tea.String(t["tName"].(string)),
				ProductKey: tea.String(a.ProductKey),
			},
		},
	}
	_, err = a.thing.BatchAddDeviceGroupRelations(batchAddDeviceGroupRelationsRequest)
	if err != nil {
		return nil, err
	}

	ret := map[string]interface{}{
		"arn":     devInfo.Body.Data.IotId,
		"pkey":    devInfo.Body.Data.DeviceSecret,
		"cert":    a.ProductKey, //accessKeyId
		"cert_id": devInfo.Body.Data.IotId,
	}

	a.logger.Debug(fmt.Sprintf("AddThing:%v", ret))
	return ret, nil
}

func (c *aliConf) load(profile string) error {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	//dir := "/tmp"
	prof, err := ioutil.ReadFile(filepath.Join(dir, "id."+profile))
	if err != nil {
		return fmt.Errorf("加载profile失败: %v", err.Error())
	}
	if err := json.Unmarshal(prof, c); err != nil {
		return fmt.Errorf("profile格式错误：%v", err.Error())
	}
	return err
}

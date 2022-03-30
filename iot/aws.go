/*
托管策略
AmazonSNSFullAccess
托管策略
AWSIoTFullAccess
*/

package iot

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	defaultMasterPolicy = "__master__"
	defaultMaster       = "__master__"
	defaultMasterGroup  = "__master__"
	defaultMasterType   = "__master__"
)

func init() {
	hostname, _ := os.Hostname()
	unCheck := []byte(defaultMaster + hostname)
	for k := range unCheck {
		switch {
		case unCheck[k] >= 'a' && unCheck[k] <= 'z':
		case unCheck[k] >= 'A' && unCheck[k] <= 'Z':
		case unCheck[k] == '_':
		case unCheck[k] == '-':
		case unCheck[k] >= '0' && unCheck[k] <= '9':
		default:
			unCheck[k] = '_'
		}
	}
	defaultMaster = string(unCheck)
}

type Aws struct {
	Options

	endpoint string
	region   string
	ca       []string

	device mqtt.Client
	thing  *iot.Client

	as *awsSaver
}

func (a *Aws) Config() map[string]interface{} {
	return map[string]interface{}{
		"region": a.region,
		//xxx.iot.ap-southeast-1.amazonaws.com
		"host": fmt.Sprintf("%s.iot.%s.amazonaws.com", a.endpoint, a.region),
		"ca":   a.ca,
	}
}

func NewAws(opt map[string]interface{}, optFn ...OptionsFunc) (Iot, error) {
	options := defaultOption()
	for _, o := range optFn {
		if err := o(&options); err != nil {
			return nil, err
		}
	}
	a := &Aws{
		Options: options,
		ca:      make([]string, 0),
		as: &awsSaver{
			MasterDevice: defaultMaster,
		},
	}

	cfg, err := config.LoadDefaultConfig(a.ctx,
		config.WithSharedConfigProfile(opt["profile"].(string)))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	a.thing = iot.NewFromConfig(cfg)

	a.region = cfg.Region
	a.endpoint = opt["endpoint"].(string)
	a.ca = amazonCa

	if err = a.createMaster(); err != nil {
		return nil, fmt.Errorf("unable to createDevice, %v", err)
	}
	return Iot(a), nil
}

func (a *Aws) Close() {
	a.cancel()
}

//Subscribe say...
//Input：
//@endpoint
//@protocol
//@topic_arn
func (a *Aws) Subscribe(s map[string]interface{}) error {
	return fmt.Errorf("no support")
}

//Publish say...
//Input：
//@message
//@topic
func (a *Aws) Publish(s map[string]interface{}) error {
	tk := a.device.Publish(s["topic"].(string), 0, s["retain"].(bool), s["message"].(string))
	a.logger.Debug("Publish:", s["message"].(string), tk)
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
func (a *Aws) CreateThing(t map[string]interface{}) (map[string]interface{}, error) {
	outThg, err := a.thing.CreateThing(a.ctx, &iot.CreateThingInput{
		ThingName:     aws.String(t["tName"].(string)),
		ThingTypeName: aws.String(t["tType"].(string)),
	})
	if err != nil {
		return nil, err
	}
	a.logger.Debug(fmt.Sprintf("CreateThing:%s", *outThg.ThingName))

	//加入组
	if _, err = a.thing.AddThingToThingGroup(a.ctx, &iot.AddThingToThingGroupInput{
		ThingName:      aws.String(t["tName"].(string)),
		ThingGroupName: aws.String(t["tGroup"].(string)),
	}); err != nil {
		return nil, err
	}

	//绑定证书
	if _, err = a.thing.AttachThingPrincipal(a.ctx, &iot.AttachThingPrincipalInput{
		ThingName: aws.String(t["tName"].(string)),
		Principal: aws.String(t["cert_arn"].(string)),
	}); err != nil {
		return nil, err
	}

	//绑定策略
	if _, err = a.thing.AttachPolicy(a.ctx, &iot.AttachPolicyInput{
		PolicyName: aws.String(t["policy"].(string)),
		Target:     aws.String(t["cert_arn"].(string)),
	}); err != nil {
		return nil, err
	}

	ret := map[string]interface{}{
		"arn": *outThg.ThingArn,
	}

	a.logger.Debug(fmt.Sprintf("AddThing:%v", ret))
	return ret, nil
}

//CreateThingAK say...
//Input：
//@tName
//Output:
//@pKey
//@certId
//@cert
//@cert_arn
func (a *Aws) CreateThingAK(ak map[string]interface{}) (map[string]interface{}, error) {
	//创建csr
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: ak["tName"].(string),
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	pKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	derBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, pKey)
	if err != nil {
		return nil, err
	}

	kOut, err := a.thing.CreateCertificateFromCsr(a.ctx, &iot.CreateCertificateFromCsrInput{
		CertificateSigningRequest: aws.String(string(pem.EncodeToMemory(
			&pem.Block{
				Type:  "CERTIFICATE REQUEST",
				Bytes: derBytes,
			}))),
		SetAsActive: true,
	})
	if err != nil {
		return nil, err
	}

	ak["pKey"] = string(pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pKey)}))
	ak["cert_id"] = *kOut.CertificateId
	ak["cert"] = *kOut.CertificatePem
	ak["cert_arn"] = *kOut.CertificateArn
	return ak, nil
}

func (a *Aws) CreateGroup(g map[string]interface{}) error {
	_, err := a.thing.CreateThingGroup(a.ctx, &iot.CreateThingGroupInput{
		ThingGroupName: aws.String(g["name"].(string)),
	})
	return err
}

func (a *Aws) CreateThingType(t map[string]interface{}) error {
	_, err := a.thing.CreateThingType(a.ctx, &iot.CreateThingTypeInput{
		ThingTypeName: aws.String(t["name"].(string)),
	})
	return err
}

//RemoveThing says ...
//@tName
//@cert_id
//@policy
func (a *Aws) RemoveThing(t map[string]interface{}) error {
	_, err := a.thing.DescribeThing(a.ctx, &iot.DescribeThingInput{
		ThingName: aws.String(t["tName"].(string)),
	})
	if err != nil {
		if !strings.Contains(err.Error(), "ResourceNotFoundException") {
			return err
		}
		return nil
	}

	x, err := a.thing.DescribeCertificate(a.ctx, &iot.DescribeCertificateInput{
		CertificateId: aws.String(t["cert_id"].(string)),
	})
	if err != nil {
		return err
	}

	_, err = a.thing.DetachThingPrincipal(a.ctx, &iot.DetachThingPrincipalInput{
		ThingName: aws.String(t["tName"].(string)),
		Principal: x.CertificateDescription.CertificateArn,
	})
	if err != nil {
		return err
	}

	if _, err = a.thing.DetachPolicy(a.ctx, &iot.DetachPolicyInput{
		PolicyName: aws.String(t["policy"].(string)),
		Target:     x.CertificateDescription.CertificateArn,
	}); err != nil {
		return err
	}

	if _, err = a.thing.DeleteCertificate(a.ctx, &iot.DeleteCertificateInput{
		CertificateId: x.CertificateDescription.CertificateId,
		ForceDelete:   true,
	}); err != nil {
		return err
	}

	return nil
}

func (a *Aws) createMaster() error {
	shouldCreate := a.as.load()
	if shouldCreate {
		a.logger.Warn("主设备[", a.as.MasterDevice, "]不存在 : 创建新主设备")
		//无设备,则创建
		nAk, err := a.CreateThingAK(map[string]interface{}{
			"tName": a.as.MasterDevice,
		})
		if err != nil {
			return err
		}
		if _, err = a.CreateThing(map[string]interface{}{
			"tName":    a.as.MasterDevice,
			"tType":    defaultMasterType,
			"tGroup":   defaultMasterGroup,
			"cert_arn": nAk["cert_arn"].(string),
			"policy":   defaultMasterPolicy,
		}); err != nil {
			return err
		}
		a.as.CertId = nAk["cert_id"].(string)
		a.as.PrivateKey = nAk["pKey"].(string)
		a.as.Cert = nAk["cert"].(string)
		if err = a.as.save(); err != nil {
			return err
		}
	}

	tlsCert, err := tls.X509KeyPair([]byte(a.as.Cert), []byte(a.as.PrivateKey))
	if err != nil {
		return fmt.Errorf("failed to load the certificates: %v", err)
	}

	certs := x509.NewCertPool()
	certs.AppendCertsFromPEM([]byte(a.ca[0]))

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		RootCAs:      certs,
	}

	if err != nil {
		return err
	}

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(fmt.Sprintf("ssl://%s:8883", a.Config()["host"].(string)))
	mqttOpts.SetMaxReconnectInterval(1 * time.Second)
	mqttOpts.SetClientID(a.as.MasterDevice)
	mqttOpts.SetTLSConfig(tlsConfig)

	c := mqtt.NewClient(mqttOpts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	a.device = c
	return nil
}

type awsSaver struct {
	MasterDevice string `json:"master_device"`
	CertId       string `json:"cert_id"`
	Cert         string `json:"cert"`
	PrivateKey   string `json:"private_key"`
}

func (as *awsSaver) save() error {
	s, err := json.Marshal(as)
	if err != nil {
		return err
	}
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return ioutil.WriteFile(filepath.Join(dir, "id.aws"), s, 0600)
}

func (as *awsSaver) load() bool {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	s, err := ioutil.ReadFile(filepath.Join(dir, as.MasterDevice+".aws"))
	if err != nil {
		return true
	}
	if err = json.Unmarshal(s, as); err != nil {
		return true
	}
	return false
}

var (
	amazonCa = []string{
		`
-----BEGIN CERTIFICATE-----
MIIDQTCCAimgAwIBAgITBmyfz5m/jAo54vB4ikPmljZbyjANBgkqhkiG9w0BAQsF
ADA5MQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRkwFwYDVQQDExBBbWF6
b24gUm9vdCBDQSAxMB4XDTE1MDUyNjAwMDAwMFoXDTM4MDExNzAwMDAwMFowOTEL
MAkGA1UEBhMCVVMxDzANBgNVBAoTBkFtYXpvbjEZMBcGA1UEAxMQQW1hem9uIFJv
b3QgQ0EgMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALJ4gHHKeNXj
ca9HgFB0fW7Y14h29Jlo91ghYPl0hAEvrAIthtOgQ3pOsqTQNroBvo3bSMgHFzZM
9O6II8c+6zf1tRn4SWiw3te5djgdYZ6k/oI2peVKVuRF4fn9tBb6dNqcmzU5L/qw
IFAGbHrQgLKm+a/sRxmPUDgH3KKHOVj4utWp+UhnMJbulHheb4mjUcAwhmahRWa6
VOujw5H5SNz/0egwLX0tdHA114gk957EWW67c4cX8jJGKLhD+rcdqsq08p8kDi1L
93FcXmn/6pUCyziKrlA4b9v7LWIbxcceVOF34GfID5yHI9Y/QCB/IIDEgEw+OyQm
jgSubJrIqg0CAwEAAaNCMEAwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMC
AYYwHQYDVR0OBBYEFIQYzIU07LwMlJQuCFmcx7IQTgoIMA0GCSqGSIb3DQEBCwUA
A4IBAQCY8jdaQZChGsV2USggNiMOruYou6r4lK5IpDB/G/wkjUu0yKGX9rbxenDI
U5PMCCjjmCXPI6T53iHTfIUJrU6adTrCC2qJeHZERxhlbI1Bjjt/msv0tadQ1wUs
N+gDS63pYaACbvXy8MWy7Vu33PqUXHeeE6V/Uq2V8viTO96LXFvKWlJbYK8U90vv
o/ufQJVtMVT8QtPHRh8jrdkPSHCa2XV4cdFyQzR1bldZwgJcJmApzyMZFo6IQ6XU
5MsI+yMRQ+hDKXJioaldXgjUkK642M4UwtBV8ob2xJNDd2ZhwLnoQdeXeGADbkpy
rqXRfboQnoZsG4q5WTP468SQvvG5
-----END CERTIFICATE-----
`,
		`
-----BEGIN CERTIFICATE-----
MIIBtjCCAVugAwIBAgITBmyf1XSXNmY/Owua2eiedgPySjAKBggqhkjOPQQDAjA5
MQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRkwFwYDVQQDExBBbWF6b24g
Um9vdCBDQSAzMB4XDTE1MDUyNjAwMDAwMFoXDTQwMDUyNjAwMDAwMFowOTELMAkG
A1UEBhMCVVMxDzANBgNVBAoTBkFtYXpvbjEZMBcGA1UEAxMQQW1hem9uIFJvb3Qg
Q0EgMzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABCmXp8ZBf8ANm+gBG1bG8lKl
ui2yEujSLtf6ycXYqm0fc4E7O5hrOXwzpcVOho6AF2hiRVd9RFgdszflZwjrZt6j
QjBAMA8GA1UdEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgGGMB0GA1UdDgQWBBSr
ttvXBp43rDCGB5Fwx5zEGbF4wDAKBggqhkjOPQQDAgNJADBGAiEA4IWSoxe3jfkr
BqWTrBqYaGFy+uGh0PsceGCmQ5nFuMQCIQCcAu/xlJyzlvnrxir4tiz+OpAUFteM
YyRIHN8wfdVoOw==
`,
	}
)

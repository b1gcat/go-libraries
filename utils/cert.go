package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

func GenerateSelfSignedCertKey(keySize int, host string, alternateIPs []net.IP, alternateDNS []string) ([]byte, []byte, error) {
	//1.生成密钥对
	priv, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}
	//2.创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1), //该号码表示CA颁发的唯一序列号，在此使用一个数来代表
		Issuer:       pkix.Name{},
		Subject:      pkix.Name{CommonName: host},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, //表示该证书是用来做服务端认证的
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}

	template.IPAddresses = append(template.IPAddresses, alternateIPs...)
	template.DNSNames = append(template.DNSNames, alternateDNS...)

	//3.创建证书,这里第二个参数和第三个参数相同则表示该证书为自签证书，返回值为DER编码的证书
	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}
	//4.将得到的证书放入pem.Block结构体中
	block := pem.Block{
		Type:    "CERTIFICATE",
		Headers: nil,
		Bytes:   certificate,
	}
	//5.通过pem编码并写入磁盘文件
	ca := bytes.NewBuffer(nil)
	pem.Encode(ca, &block)

	//6.将私钥中的密钥对放入pem.Block结构体中
	block = pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(priv),
	}
	caKey := bytes.NewBuffer(nil)
	pem.Encode(caKey, &block)

	return ca.Bytes(), caKey.Bytes(), nil
}

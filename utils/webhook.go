package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

/*
	/*
		curl -XPOST -H 'Content-Type: application/json' "https://oapi.dingtalk.com/robot/send?access_token=xxxx" -d '{
		        "msgtype": "markdown",
		        "markdown": {
		        		"title":"Alert",
		            "text": "test"
		        }
		}'
*/
//DingDing says ...
func DingDing(webHook string, aJson []byte) error {
	tr := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(5000) * time.Millisecond,
			KeepAlive: -1,
		}).DialContext,
		ResponseHeaderTimeout: time.Duration(5000) * time.Millisecond,
		TLSHandshakeTimeout:   time.Duration(5000) * time.Millisecond,
	}

	cli := &http.Client{
		Transport: tr,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("forbidden redirects(10)")
			}
			return nil
		},
	}
	request, err := http.NewRequest("POST", webHook, bytes.NewReader(aJson))
	if err != nil {
		return err
	}
	request.Header.Add("Accept-Encoding", "gzip")
	request.Header.Add("Content-type", "application/json")
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")
	response, _ := cli.Do(request)
	if response != nil {
		defer response.Body.Close()
	}
	return nil
}

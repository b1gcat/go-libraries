package iot

import (
	"context"
	"fmt"
	"testing"
)

func Test_aliYun(t *testing.T) {
	a, err := NewAli(map[string]interface{}{
		"profile":  "aliYun",
		"endpoint": "ap-southeast-1",
	}, WithContext(context.Background()))
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}

	fmt.Println(a.Publish(map[string]interface{}{
		"topic":   "R110",
		"message": `{"message":"fucker"}`,
		"retain":  false,
	}))
}

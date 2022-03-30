package iot

import (
	"context"
	"fmt"
	"testing"
)

func Test_aws(t *testing.T) {
	a, err := NewAws(map[string]interface{}{
		"profile":  "aws",
		"endpoint": "xxx",
	}, WithContext(context.Background()))
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}

	fmt.Println(a.Publish(map[string]interface{}{
		"topic":   "__master__b1gcat_local",
		"message": `{"message":"hello"}`,
		"retain":  false,
	}))
}

func Test_delete(t *testing.T) {
	a, err := NewAws(map[string]interface{}{
		"profile":  "aws",
		"endpoint": "xxxxxxxx",
	}, WithContext(context.Background()))
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}
	err = a.(*Aws).RemoveThing(map[string]interface{}{
		"tName":   "test",
		"policy":  "unknown",
		"cert_id": "123",
	})
	fmt.Println(err.Error())
}

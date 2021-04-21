package redis

import (
	"encoding/json"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"

	"github.com/crossplane/provider-alibaba/apis/redis/v1alpha1"
)

func TestGenerateObservation(t *testing.T) {
	ob := GenerateObservation(&DBInstance{Status: v1alpha1.RedisInstanceStateRunning})
	if ob.DBInstanceStatus != v1alpha1.RedisInstanceStateRunning {
		t.Errorf("RedisInstanceStatus: want=%v, get=%v", v1alpha1.RedisInstanceStateRunning, ob.DBInstanceStatus)
	}
}

func TestIsErrorNotFound(t *testing.T) {
	var response = make(map[string]string)
	response["Code"] = "InvalidInstanceId.NotFound"

	responseContent, _ := json.Marshal(response)
	err := errors.NewServerError(404, string(responseContent), "comment")
	isErrorNotFound := IsErrorNotFound(err)
	if !isErrorNotFound {
		t.Errorf("IsErrorNotFound: want=%v, get=%v", true, isErrorNotFound)
	}
}

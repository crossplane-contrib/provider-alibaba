package redis

import (
	"encoding/json"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	aliredis "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"

	"github.com/crossplane-contrib/provider-alibaba/apis/redis/v1alpha1"
)

func TestGenerateObservation(t *testing.T) {
	ob := GenerateObservation(&aliredis.DBInstanceAttribute{
		InstanceStatus:   v1alpha1.RedisInstanceStateRunning,
		InstanceId:       "test-id",
		ConnectionDomain: "test-address",
		Port:             8080,
	})
	if ob.InstanceStatus != v1alpha1.RedisInstanceStateRunning {
		t.Errorf("RedisInstanceStatus: want=%v, get=%v", v1alpha1.RedisInstanceStateRunning, ob.InstanceStatus)
	}
}

func TestIsErrorNotFound(t *testing.T) {
	var response = make(map[string]string)
	response["Code"] = errInstanceNotFoundCode

	responseContent, _ := json.Marshal(response) //nolint:errchkjson
	err := errors.NewServerError(404, string(responseContent), "comment")
	isErrorNotFound := IsErrorNotFound(err)
	if !isErrorNotFound {
		t.Errorf("IsErrorNotFound: want=%v, get=%v", true, isErrorNotFound)
	}
}

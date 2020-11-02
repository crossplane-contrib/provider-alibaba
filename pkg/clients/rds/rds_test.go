package rds

import (
	"encoding/json"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"

	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
)

func TestGenerateObservation(t *testing.T) {
	ob := GenerateObservation(&DBInstance{Status: v1alpha1.RDSInstanceStateRunning})
	if ob.DBInstanceStatus != v1alpha1.RDSInstanceStateRunning {
		t.Errorf("DBInstanceStatus: want=%v, get=%v", v1alpha1.RDSInstanceStateRunning, ob.DBInstanceStatus)
	}
}

func TestIsErrorNotFound(t *testing.T) {
	var response = make(map[string]string)
	response["Code"] = ErrCodeInstanceNotFound

	responseContent, _ := json.Marshal(response)
	err := errors.NewServerError(404, string(responseContent), "comment")
	isErrorNotFound := IsErrorNotFound(err)
	if !isErrorNotFound {
		t.Errorf("IsErrorNotFound: want=%v, get=%v", true, isErrorNotFound)
	}
}

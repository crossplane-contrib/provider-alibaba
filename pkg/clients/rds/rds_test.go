package rds

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/stretchr/testify/assert"

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

func TestDBInstanceOperations(t *testing.T) {
	ctx := context.TODO()
	c, err := NewClient(ctx, "abc", "def", "cn-beijing")
	assert.Nil(t, err)

	db, err := c.DescribeDBInstance("1")
	assert.Nil(t, db)
	assert.Error(t, err)

	req := CreateDBInstanceRequest{
		Name:                  "abc",
		Engine:                "mysql",
		EngineVersion:         "8.0",
		DBInstanceClass:       "big",
		DBInstanceStorageInGB: 20,
		SecurityIPList:        "1.2.3.0/24",
	}
	db, err = c.CreateDBInstance(&req)
	assert.Nil(t, db)
	assert.Error(t, err)

	err = c.DeleteDBInstance("1")
	assert.Error(t, err)

	err = c.CreateAccount("1", "operator", "ABC123")
	assert.Error(t, err)
}

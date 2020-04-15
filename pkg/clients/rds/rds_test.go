package rds

import (
	"testing"

	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
)

func TestGenerateObservation(t *testing.T) {
	ob := GenerateObservation(&DBInstance{Status: v1alpha1.RDSInstanceStateRunning})
	if ob.DBInstanceStatus != v1alpha1.RDSInstanceStateRunning {
		t.Errorf("DBInstanceStatus: want=%v, get=%v", v1alpha1.RDSInstanceStateRunning, ob.DBInstanceStatus)
	}
}

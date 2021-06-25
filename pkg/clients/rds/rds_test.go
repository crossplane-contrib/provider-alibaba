package rds

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
)

func TestGenerateObservation(t *testing.T) {
	type args struct {
		dbInstance *DBInstance
	}
	type want struct {
		dbStatus string
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SuccessfullyGenerateObservation": {
			args: args{
				dbInstance: &DBInstance{Status: v1alpha1.RDSInstanceStateRunning},
			},
			want: want{
				dbStatus: v1alpha1.RDSInstanceStateRunning,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dbInstance := GenerateObservation(tc.args.dbInstance)
			if diff := cmp.Diff(tc.want.dbStatus, dbInstance.DBInstanceStatus, test.EquateConditions()); diff != "" {
				t.Errorf("\nGenerateObservation(...) %s\n", diff)
			}
		})
	}
}

func TestIsErrorNotFound(t *testing.T) {
	var response = make(map[string]string)
	response["Code"] = ErrCodeInstanceNotFound

	responseContent, _ := json.Marshal(response)

	type args struct {
		httpStatus      int
		responseContent string
		comment         string
	}
	type want struct {
		found bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"ErrorNotFound": {
			args: args{
				httpStatus:      404,
				responseContent: string(responseContent),
				comment:         "comment",
			},
			want: want{
				found: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := errors.NewServerError(tc.args.httpStatus, tc.args.responseContent, tc.args.comment)
			isErrorNotFound := IsErrorNotFound(err)
			if diff := cmp.Diff(tc.want.found, isErrorNotFound, test.EquateConditions()); diff != "" {
				t.Errorf("\nIsErrorNotFound(...) %s\n", diff)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	ctx := context.TODO()
	type args struct {
		ctx             context.Context
		accessKeyID     string
		accessKeySecret string
		securityToken   string
		region          string
	}
	type want struct {
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FailedToNewClient": {
			args: args{
				ctx:             ctx,
				accessKeyID:     "xwerwrfYfwq934tsfsFAKED",
				accessKeySecret: "fsdfwerfaUIIffaYYYYYFUUFHUDSDKSDFAKED",
				region:          "cn-beijing",
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := NewClient(tc.args.ctx, tc.args.accessKeyID, tc.args.accessKeySecret, tc.args.securityToken, tc.args.region)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\nNewClient(...) -want error, +got error:\n%s\n", diff)
			}
		})
	}
}

func TestDescribeDBInstance(t *testing.T) {
	c, _ := NewClient(context.TODO(), "xwerwrfYfwq934tsfsFAKED", "fsdfwerfaUIIffaYYYYYFUUFHUDSDKSDFAKED", "", "cn-beijing")
	type args struct {
		id string
	}

	type want struct {
		obj     *DBInstance
		errCode string
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FailedToDescribeDBInstance": {
			args: args{
				id: "1",
			},
			want: want{
				obj:     nil,
				errCode: "InvalidAccessKeyId.NotFound",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			db, err := c.DescribeDBInstance(tc.args.id)
			if err != nil {
				e, ok := err.(*errors.ServerError)
				if ok {
					if diff := cmp.Diff(tc.want.errCode, e.ErrorCode(), test.EquateConditions()); diff != "" {
						t.Errorf("\nDescribeDBInstance(...) -want errorCode, +got errorCode:\n%s\n", diff)
					}
				}
			}

			if diff := cmp.Diff(tc.want.obj, db, test.EquateConditions()); diff != "" {
				t.Errorf("\nDescribeDBInstance(...) %s\n", diff)
			}
		})
	}
}

func TestCreateDBInstance(t *testing.T) {
	c, _ := NewClient(context.TODO(), "xwerwrfYfwq934tsfsFAKED", "fsdfwerfaUIIffaYYYYYFUUFHUDSDKSDFAKED", "", "cn-beijing")
	type args struct {
		req CreateDBInstanceRequest
	}

	type want struct {
		obj     *DBInstance
		errCode string
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FailedToCreateDBInstance": {
			args: args{
				req: CreateDBInstanceRequest{
					Name:                  "abc",
					Engine:                "mysql",
					EngineVersion:         "8.0",
					DBInstanceClass:       "big",
					DBInstanceStorageInGB: 20,
					SecurityIPList:        "1.2.3.0/24",
				},
			},
			want: want{
				obj:     nil,
				errCode: "InvalidAccessKeyId.NotFound",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			db, err := c.CreateDBInstance(&tc.args.req)
			if err != nil {
				e, ok := err.(*errors.ServerError)
				if ok {
					if diff := cmp.Diff(tc.want.errCode, e.ErrorCode(), test.EquateConditions()); diff != "" {
						t.Errorf("\nCreateDBInstance(...) -want errorCode, +got errorCode:\n%s\n", diff)
					}
				}
			}

			if diff := cmp.Diff(tc.want.obj, db, test.EquateConditions()); diff != "" {
				t.Errorf("\nCreateDBInstance(...) %s\n", diff)
			}
		})
	}
}

func TestDeleteDBInstance(t *testing.T) {
	c, _ := NewClient(context.TODO(), "xwerwrfYfwq934tsfsFAKED", "fsdfwerfaUIIffaYYYYYFUUFHUDSDKSDFAKED", "", "cn-beijing")
	type args struct {
		id string
	}

	type want struct {
		errCode string
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FailedToDeleteDBInstance": {
			args: args{
				id: "123",
			},
			want: want{
				errCode: "InvalidAccessKeyId.NotFound",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := c.DeleteDBInstance(tc.args.id)
			if err != nil {
				e, ok := err.(*errors.ServerError)
				if ok {
					if diff := cmp.Diff(tc.want.errCode, e.ErrorCode(), test.EquateConditions()); diff != "" {
						t.Errorf("\nDeleteDBInstance(...) -want errorCode, +got errorCode:\n%s\n", diff)
					}
				}
			}
		})
	}
}

func TestCreateAccount(t *testing.T) {
	c, _ := NewClient(context.TODO(), "xwerwrfYfwq934tsfsFAKED", "fsdfwerfaUIIffaYYYYYFUUFHUDSDKSDFAKED", "", "cn-beijing")
	type args struct {
		id       string
		username string
		password string
	}

	type want struct {
		errCode string
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FailedToCreateAccount": {
			args: args{
				id:       "123",
				username: "operator",
				password: "ABC123",
			},
			want: want{
				errCode: "InvalidAccessKeyId.NotFound",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := c.CreateAccount(tc.args.id, tc.args.username, tc.args.password)
			if err != nil {
				e, ok := err.(*errors.ServerError)
				if ok {
					if diff := cmp.Diff(tc.want.errCode, e.ErrorCode(), test.EquateConditions()); diff != "" {
						t.Errorf("\nCreateAccount(...) -want errorCode, +got errorCode:\n%s\n", diff)
					}
				}
			}
		})
	}
}

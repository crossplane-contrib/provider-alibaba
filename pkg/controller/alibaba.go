/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/provider-alibaba/pkg/controller/config"
	"github.com/crossplane/provider-alibaba/pkg/controller/database"
	"github.com/crossplane/provider-alibaba/pkg/controller/database/redis"
	"github.com/crossplane/provider-alibaba/pkg/controller/nas"
	"github.com/crossplane/provider-alibaba/pkg/controller/oss"
	"github.com/crossplane/provider-alibaba/pkg/controller/slb"
	"github.com/crossplane/provider-alibaba/pkg/controller/sls"
)

// Setup creates Alibaba controllers with the supplied logger and adds them to the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	for _, setup := range []func(ctrl.Manager, logging.Logger) error{
		config.Setup,
		database.SetupRDSInstance,
		redis.SetupRedisInstance,
		sls.SetupProject,
		sls.SetupStore,
		oss.SetupBucket,
		nas.SetupNASFileSystem,
		nas.SetupNASMountTarget,
		slb.SetupCLB,
	} {
		if err := setup(mgr, l); err != nil {
			return err
		}
	}
	return nil
}

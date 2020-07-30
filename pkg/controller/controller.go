/*
Copyright 2016 The Fission Authors.

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
	"context"
	"go.uber.org/zap"

	"github.com/fission/fission/pkg/crd"
)

// jingtao-note: Controller启动入口
func Start(logger *zap.Logger, port int, unitTestFlag bool) {
	cLogger := logger.Named("controller")

	//  jingtao-note: 获取fission客户端、k8s的客户端和k8s api，其中k8s支持使用绑定的服务账号进行获取
	fc, kc, apiExtClient, err := crd.MakeFissionClient()
	if err != nil {
		cLogger.Fatal("failed to connect to k8s API", zap.Error(err))
	}
	// jingtao-note: 保证Fission自定义资源的存在
	err = crd.EnsureFissionCRDs(cLogger, apiExtClient)
	if err != nil {
		cLogger.Fatal("failed to create fission CRDs", zap.Error(err))
	}
	// jingtao-note: 调用了一个自定义资源的操作函数，确保所有的自定义资源都在K8s上配置完成
	err = fc.WaitForCRDs()
	if err != nil {
		cLogger.Fatal("error waiting for CRDs", zap.Error(err))
	}
	// jingtao-note: 了解context的原理与应用，感觉是回滚操作使用的
	ctx, cancel := context.WithCancel(context.Background())
	featureStatus, err := ConfigureFeatures(ctx, cLogger, unitTestFlag, fc, kc)
	if err != nil {
		cLogger.Error("error configuring features - proceeding without optional features", zap.Error(err))
	}
	defer cancel()

	api, err := MakeAPI(cLogger, featureStatus)
	if err != nil {
		cLogger.Fatal("failed to start controller", zap.Error(err))
	}
	api.Serve(port)
}

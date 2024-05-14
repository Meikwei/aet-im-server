// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discoveryregister

import (
	"time"

	"github.com/Meikwei/go-tools/discovery"
	"github.com/Meikwei/go-tools/discovery/zookeeper"
	"github.com/Meikwei/go-tools/errs"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/open-im-server/v3/pkg/common/discoveryregister/kubernetes"
)

const (
	zookeeperConst = "zookeeper"
	kubenetesConst = "k8s"
	directConst    = "direct"
)

// NewDiscoveryRegister 创建一个基于提供的环境类型的服务发现和注册客户端。
//
// 参数:
// - zookeeperConfig: 包含ZooKeeper配置信息的结构体指针。
// - share: 包含共享配置信息的结构体指针。
//
// 返回值:
// - discovery.SvcDiscoveryRegistry: 实现了服务发现和注册接口的客户端实例。
// - error: 在创建过程中遇到错误时返回的错误信息。
func NewDiscoveryRegister(zookeeperConfig *config.ZooKeeper, share *config.Share) (discovery.SvcDiscoveryRegistry, error) {
	switch share.Env {
	case zookeeperConst:
		// 创建一个ZooKeeper客户端
		return zookeeper.NewZkClient(
			zookeeperConfig.Address,
			zookeeperConfig.Schema,
			zookeeper.WithFreq(time.Hour),
			zookeeper.WithUserNameAndPassword(zookeeperConfig.Username, zookeeperConfig.Password),
			zookeeper.WithRoundRobin(),
			zookeeper.WithTimeout(10),
		)
	case kubenetesConst:
		// 创建一个Kubernetes服务发现注册客户端
		return kubernetes.NewK8sDiscoveryRegister(share.RpcRegisterName.MessageGateway)
	case directConst:
		// 创建一个直接连接客户端（代码未完整展示）
		//return direct.NewConnDirect(config)
	default:
		// 对于不支持的服务发现类型，返回错误
		return nil, errs.New("unsupported discovery type", "type", share.Env).Wrap()
	}
	return nil, nil
}

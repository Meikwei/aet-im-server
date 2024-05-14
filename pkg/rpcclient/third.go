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

package rpcclient

import (
	"context"

	"github.com/Meikwei/go-tools/discovery"
	"github.com/Meikwei/go-tools/system/program"
	"github.com/openimsdk/protocol/third"
	"google.golang.org/grpc"
)

type Third struct {
	conn       grpc.ClientConnInterface
	Client     third.ThirdClient
	discov     discovery.SvcDiscoveryRegistry
	GrafanaUrl string
}

// NewThird 创建一个新的Third实例。
//
// 参数:
// - discov: 一个实现了SvcDiscoveryRegistry接口的对象，用于服务发现。
// - rpcRegisterName: 用于注册或标识Third服务的名称。
// - grafanaUrl: Grafana的URL，用于访问Grafana服务。
//
// 返回值:
// - 返回一个初始化好的*Third对象。
func NewThird(discov discovery.SvcDiscoveryRegistry, rpcRegisterName, grafanaUrl string) *Third {
    // 通过服务发现获取与Third服务的连接。
	conn, err := discov.GetConn(context.Background(), rpcRegisterName)
	if err != nil {
		program.ExitWithError(err) // 如果获取连接失败，则退出程序。
	}
	
    // 使用获取的连接创建Third服务的客户端。
	client := third.NewThirdClient(conn)
	if err != nil {
		program.ExitWithError(err) // 如果创建客户端失败，则退出程序。
	}
	
    // 返回初始化好的Third实例。
	return &Third{discov: discov, Client: client, conn: conn, GrafanaUrl: grafanaUrl}
}

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

package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Meikwei/go-tools/utils/datautil"
	"github.com/Meikwei/go-tools/utils/network"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"

	"github.com/Meikwei/go-tools/discovery"
	"github.com/Meikwei/go-tools/errs"
	"github.com/Meikwei/go-tools/log"
	"github.com/Meikwei/go-tools/system/program"
	kdisc "github.com/openimsdk/open-im-server/v3/pkg/common/discoveryregister"
	ginprom "github.com/openimsdk/open-im-server/v3/pkg/common/ginprometheus"
	"github.com/openimsdk/open-im-server/v3/pkg/common/prommetrics"
)

type Config struct {
	RpcConfig          config.API
	MongodbConfig      config.Mongo
	ZookeeperConfig    config.ZooKeeper
	NotificationConfig config.Notification
	Share              config.Share
	MinioConfig        config.Minio
}

// Start 初始化并启动服务
// ctx: 上下文，用于控制程序运行期间的流程。
// index: 用于指定配置中端口的索引。
// config: 启动配置参数。
// 返回值: 如果启动过程中遇到错误，返回错误信息。
func Start(ctx context.Context, index int, config *Config) error {
    // 根据索引获取API和Prometheus端口
	apiPort, err := datautil.GetElemByIndex(config.RpcConfig.Api.Ports, index)
	if err != nil {
		return err
	}
	prometheusPort, err := datautil.GetElemByIndex(config.RpcConfig.Prometheus.Ports, index)
	if err != nil {
		return err
	}

	var client discovery.SvcDiscoveryRegistry

    // 根据是否是集群部署来决定是否使用zk注册中心
	client, err = kdisc.NewDiscoveryRegister(&config.ZookeeperConfig, &config.Share)
	if err != nil {
		return errs.WrapMsg(err, "failed to register discovery service")
	}

	var (
		netDone = make(chan struct{}, 1) // 用于通知网络服务启动完成或出错
		netErr  error
	)

	router := newGinRouter(client, config) // 创建Gin路由器

    // 如果启用了Prometheus监控，则另起一个goroutine启动Prometheus监听
	if config.RpcConfig.Prometheus.Enable {
		go func() {
			p := ginprom.NewPrometheus("app", prommetrics.GetGinCusMetrics("Api"))
			p.SetListenAddress(fmt.Sprintf(":%d", prometheusPort))
			if err = p.Use(router); err != nil && err != http.ErrServerClosed {
				netErr = errs.WrapMsg(err, fmt.Sprintf("prometheus start err: %d", prometheusPort))
				netDone <- struct{}{}
			}
		}()
	}

    // 创建API服务并指定监听地址和路由器
	address := net.JoinHostPort(network.GetListenIP(config.RpcConfig.Api.ListenIP), strconv.Itoa(apiPort))
	server := http.Server{Addr: address, Handler: router}
	log.CInfo(ctx, "API server is initializing", "address", address, "apiPort", apiPort, "prometheusPort", prometheusPort)
	go func() {
		err = server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			netErr = errs.WrapMsg(err, fmt.Sprintf("api start err: %s", server.Addr))
			netDone <- struct{}{}
		}
	}()

    // 监听系统信号，用于优雅关闭服务
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // 设置超时时间，确保能够及时关闭
	defer cancel()
	select {
	case <-sigs:
		program.SIGTERMExit() // 处理SIGTERM信号，进行优雅关闭
		err := server.Shutdown(ctx)
		if err != nil {
			return errs.WrapMsg(err, "shutdown err")
		}
	case <-netDone:
		close(netDone)
		return netErr // 返回网络启动错误
	}
	return nil
}

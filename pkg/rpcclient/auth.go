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
	"github.com/Meikwei/protocol/auth"
	pbAuth "github.com/Meikwei/protocol/auth"
	"google.golang.org/grpc"
)

// NewAuth 创建一个新的Auth实例。
//
// 参数:
// discov - 服务发现注册接口，用于获取与RPC服务的连接。
// rpcRegisterName - RPC服务的注册名称，用于标识特定的服务实例。
//
// 返回值:
// 返回一个初始化好的Auth指针。
func NewAuth(discov discovery.SvcDiscoveryRegistry, rpcRegisterName string) *Auth {
    // 通过服务发现获取与指定RPC服务的连接
	conn, err := discov.GetConn(context.Background(), rpcRegisterName)
	if err != nil {
		// 如果获取连接失败，则程序异常退出
		program.ExitWithError(err)
	}
	// 基于获取的连接，创建Auth客户端实例
	client := auth.NewAuthClient(conn)
	return &Auth{discov: discov, conn: conn, Client: client}
}

type Auth struct {
	conn   grpc.ClientConnInterface
	Client auth.AuthClient
	discov discovery.SvcDiscoveryRegistry
}

func (a *Auth) ParseToken(ctx context.Context, token string) (*pbAuth.ParseTokenResp, error) {
	req := pbAuth.ParseTokenReq{
		Token: token,
	}
	resp, err := a.Client.ParseToken(ctx, &req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (a *Auth) InvalidateToken(ctx context.Context, preservedToken, userID string, platformID int) (*pbAuth.InvalidateTokenResp, error) {
	req := pbAuth.InvalidateTokenReq{
		PreservedToken: preservedToken,
		UserID:         userID,
		PlatformID:     int32(platformID),
	}
	resp, err := a.Client.InvalidateToken(ctx, &req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

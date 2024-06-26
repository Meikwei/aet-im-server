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
	"github.com/Meikwei/protocol/friend"
	sdkws "github.com/Meikwei/protocol/sdkws"
	"google.golang.org/grpc"
)

type Friend struct {
	conn   grpc.ClientConnInterface
	Client friend.FriendClient
	discov discovery.SvcDiscoveryRegistry
}

// NewFriend 创建一个 Friend 实例。
// 
// param discov: 服务发现注册接口，用于获取远程服务的连接。
// param rpcRegisterName: RPC服务的注册名称，用于标识具体的服务。
// return: 返回初始化好的 Friend 指针。
func NewFriend(discov discovery.SvcDiscoveryRegistry, rpcRegisterName string) *Friend {
    // 通过服务发现获取与指定RPC服务的连接
	conn, err := discov.GetConn(context.Background(), rpcRegisterName)
	if err != nil {
		// 如果获取连接失败，则退出程序
		program.ExitWithError(err)
	}
	// 根据连接创建 Friend 客户端
	client := friend.NewFriendClient(conn)
	return &Friend{discov: discov, conn: conn, Client: client}
}

type FriendRpcClient Friend

func NewFriendRpcClient(discov discovery.SvcDiscoveryRegistry, rpcRegisterName string) FriendRpcClient {
	return FriendRpcClient(*NewFriend(discov, rpcRegisterName))
}

func (f *FriendRpcClient) GetFriendsInfo(
	ctx context.Context,
	ownerUserID, friendUserID string,
) (resp *sdkws.FriendInfo, err error) {
	r, err := f.Client.GetDesignatedFriends(
		ctx,
		&friend.GetDesignatedFriendsReq{OwnerUserID: ownerUserID, FriendUserIDs: []string{friendUserID}},
	)
	if err != nil {
		return nil, err
	}
	resp = r.FriendsInfo[0]
	return
}

// possibleFriendUserID Is PossibleFriendUserId's friends.
func (f *FriendRpcClient) IsFriend(ctx context.Context, possibleFriendUserID, userID string) (bool, error) {
	resp, err := f.Client.IsFriend(ctx, &friend.IsFriendReq{UserID1: userID, UserID2: possibleFriendUserID})
	if err != nil {
		return false, err
	}
	return resp.InUser1Friends, nil
}

func (f *FriendRpcClient) GetFriendIDs(ctx context.Context, ownerUserID string) (friendIDs []string, err error) {
	req := friend.GetFriendIDsReq{UserID: ownerUserID}
	resp, err := f.Client.GetFriendIDs(ctx, &req)
	if err != nil {
		return nil, err
	}
	return resp.FriendIDs, nil
}

func (b *FriendRpcClient) IsBlack(ctx context.Context, possibleBlackUserID, userID string) (bool, error) {
	r, err := b.Client.IsBlack(ctx, &friend.IsBlackReq{UserID1: possibleBlackUserID, UserID2: userID})
	if err != nil {
		return false, err
	}
	return r.InUser2Blacks, nil
}

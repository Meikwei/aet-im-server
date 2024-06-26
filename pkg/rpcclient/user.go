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
	"strings"

	"github.com/Meikwei/go-tools/discovery"
	"github.com/Meikwei/go-tools/system/program"
	"github.com/Meikwei/go-tools/utils/datautil"
	"github.com/Meikwei/protocol/sdkws"
	"github.com/Meikwei/protocol/user"
	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/common/servererrs"
	"google.golang.org/grpc"
)

// User represents a structure holding connection details for the User RPC client.
type User struct {
	conn                  grpc.ClientConnInterface
	Client                user.UserClient
	Discov                discovery.SvcDiscoveryRegistry
	MessageGateWayRpcName string
	imAdminUserID         []string
}

// NewUser 使用提供的服务发现注册表初始化并返回一个新的User实例。
// 它通过注册表连接到UserService，创建用于用户操作的客户端，
// 然后返回指向新创建的User实例的指针。
//
// 参数:
// - discov: 一个discovery.SvcDiscoveryRegistry实例，用于发现并连接到UserService。
// - rpcRegisterName: 用于注册或发现的RPC服务名称。用于标识UserService。
// - messageGateWayRpcName: 消息网关的RPC服务名称。用于与消息网关通信。
// - imAdminUserID: 一个字符串切片，表示管理员用户的ID。
//
// 返回:
// - 初始化后的User实例的指针。
func NewUser(discov discovery.SvcDiscoveryRegistry, rpcRegisterName, messageGateWayRpcName string,
	imAdminUserID []string) *User {
	// 尝试使用服务发现注册表获取与UserService的连接。
	conn, err := discov.GetConn(context.Background(), rpcRegisterName)
	if err != nil {
		// 如果尝试连接时发生错误，程序将退出。
		program.ExitWithError(err)
	}
	// 创建一个用于用户操作的新客户端。
	client := user.NewUserClient(conn)
	// 初始化并返回User实例。
	return &User{Discov: discov, Client: client,
		conn:                  conn,
		MessageGateWayRpcName: messageGateWayRpcName,
		imAdminUserID:         imAdminUserID}
}

// UserRpcClient represents the structure for a User RPC client.
type UserRpcClient User

// NewUserRpcClientByUser initializes a UserRpcClient based on the provided User instance.
func NewUserRpcClientByUser(user *User) *UserRpcClient {
	rpc := UserRpcClient(*user)
	return &rpc
}

// NewUserRpcClient initializes a UserRpcClient based on the provided service discovery registry.
func NewUserRpcClient(client discovery.SvcDiscoveryRegistry, rpcRegisterName string,
	imAdminUserID []string) UserRpcClient {
	return UserRpcClient(*NewUser(client, rpcRegisterName, "", imAdminUserID))
}

// GetUsersInfo retrieves information for multiple users based on their user IDs.
func (u *UserRpcClient) GetUsersInfo(ctx context.Context, userIDs []string) ([]*sdkws.UserInfo, error) {
	if len(userIDs) == 0 {
		return []*sdkws.UserInfo{}, nil
	}
	resp, err := u.Client.GetDesignateUsers(ctx, &user.GetDesignateUsersReq{
		UserIDs: userIDs,
	})
	if err != nil {
		return nil, err
	}
	if ids := datautil.Single(userIDs, datautil.Slice(resp.UsersInfo, func(e *sdkws.UserInfo) string {
		return e.UserID
	})); len(ids) > 0 {
		return nil, servererrs.ErrUserIDNotFound.WrapMsg(strings.Join(ids, ","))
	}
	return resp.UsersInfo, nil
}

// GetUserInfo retrieves information for a single user based on the provided user ID.
func (u *UserRpcClient) GetUserInfo(ctx context.Context, userID string) (*sdkws.UserInfo, error) {
	users, err := u.GetUsersInfo(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	return users[0], nil
}

// GetUsersInfoMap retrieves a map of user information indexed by their user IDs.
func (u *UserRpcClient) GetUsersInfoMap(ctx context.Context, userIDs []string) (map[string]*sdkws.UserInfo, error) {
	users, err := u.GetUsersInfo(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	return datautil.SliceToMap(users, func(e *sdkws.UserInfo) string {
		return e.UserID
	}), nil
}

// GetPublicUserInfos retrieves public information for multiple users based on their user IDs.
func (u *UserRpcClient) GetPublicUserInfos(
	ctx context.Context,
	userIDs []string,
	complete bool,
) ([]*sdkws.PublicUserInfo, error) {
	users, err := u.GetUsersInfo(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	return datautil.Slice(users, func(e *sdkws.UserInfo) *sdkws.PublicUserInfo {
		return &sdkws.PublicUserInfo{
			UserID:   e.UserID,
			Nickname: e.Nickname,
			FaceURL:  e.FaceURL,
			Ex:       e.Ex,
		}
	}), nil
}

// GetPublicUserInfo retrieves public information for a single user based on the provided user ID.
func (u *UserRpcClient) GetPublicUserInfo(ctx context.Context, userID string) (*sdkws.PublicUserInfo, error) {
	users, err := u.GetPublicUserInfos(ctx, []string{userID}, true)
	if err != nil {
		return nil, err
	}
	return users[0], nil
}

// GetPublicUserInfoMap retrieves a map of public user information indexed by their user IDs.
func (u *UserRpcClient) GetPublicUserInfoMap(
	ctx context.Context,
	userIDs []string,
	complete bool,
) (map[string]*sdkws.PublicUserInfo, error) {
	users, err := u.GetPublicUserInfos(ctx, userIDs, complete)
	if err != nil {
		return nil, err
	}
	return datautil.SliceToMap(users, func(e *sdkws.PublicUserInfo) string {
		return e.UserID
	}), nil
}

// GetUserGlobalMsgRecvOpt retrieves the global message receive option for a user based on the provided user ID.
func (u *UserRpcClient) GetUserGlobalMsgRecvOpt(ctx context.Context, userID string) (int32, error) {
	resp, err := u.Client.GetGlobalRecvMessageOpt(ctx, &user.GetGlobalRecvMessageOptReq{
		UserID: userID,
	})
	if err != nil {
		return 0, err
	}
	return resp.GlobalRecvMsgOpt, nil
}

// Access verifies the access rights for the provided user ID.
func (u *UserRpcClient) Access(ctx context.Context, ownerUserID string) error {
	_, err := u.GetUserInfo(ctx, ownerUserID)
	if err != nil {
		return err
	}
	return authverify.CheckAccessV3(ctx, ownerUserID, u.imAdminUserID)
}

// GetAllUserIDs retrieves all user IDs with pagination options.
func (u *UserRpcClient) GetAllUserIDs(ctx context.Context, pageNumber, showNumber int32) ([]string, error) {
	resp, err := u.Client.GetAllUserID(ctx, &user.GetAllUserIDReq{Pagination: &sdkws.RequestPagination{PageNumber: pageNumber, ShowNumber: showNumber}})
	if err != nil {
		return nil, err
	}
	return resp.UserIDs, nil
}

// SetUserStatus sets the status for a user based on the provided user ID, status, and platform ID.
func (u *UserRpcClient) SetUserStatus(ctx context.Context, userID string, status int32, platformID int) error {
	_, err := u.Client.SetUserStatus(ctx, &user.SetUserStatusReq{
		UserID: userID,
		Status: status, PlatformID: int32(platformID),
	})
	return err
}

func (u *UserRpcClient) GetNotificationByID(ctx context.Context, userID string) error {
	_, err := u.Client.GetNotificationAccount(ctx, &user.GetNotificationAccountReq{
		UserID: userID,
	})
	return err
}

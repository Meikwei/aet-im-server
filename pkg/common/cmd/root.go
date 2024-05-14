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

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Meikwei/go-tools/errs"
	"github.com/Meikwei/go-tools/log"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/spf13/cobra"
)

// RootCmd 结构体定义了根命令的基本配置和状态
type RootCmd struct {
	Command        cobra.Command // Command 代表了这个根命令的cobra结构体
	processName    string        // processName 定义了进程的名称
	port           int           // port 指定了服务监听的端口
	prometheusPort int           // prometheusPort 指定了Prometheus监控的端口
	log            config.Log    // log 包含了日志配置信息
	index          int           // index 用于标识或索引RootCmd实例
}

// Index 返回RootCmd实例的索引号
func (r *RootCmd) Index() int {
	return r.index
}

// Port 返回服务监听的端口号
func (r *RootCmd) Port() int {
	return r.port
}

// CmdOpts 定义了命令行选项的结构体，包括日志前缀名和配置映射。
type CmdOpts struct {
	loggerPrefixName string
	configMap        map[string]any
}

// WithCronTaskLogName 返回一个函数，该函数设置CmdOpts的日志前缀名为"openim-crontask"。
func WithCronTaskLogName() func(*CmdOpts) {
	return func(opts *CmdOpts) {
		opts.loggerPrefixName = "openim-crontask"
	}
}

// WithLogName 返回一个函数，该函数根据传入的logName设置CmdOpts的日志前缀名。
func WithLogName(logName string) func(*CmdOpts) {
	return func(opts *CmdOpts) {
		opts.loggerPrefixName = logName
	}
}

// WithConfigMap 返回一个函数，该函数根据传入的configMap设置CmdOpts的配置映射。
func WithConfigMap(configMap map[string]any) func(*CmdOpts) {
	return func(opts *CmdOpts) {
		opts.configMap = configMap
	}
}

// NewRootCmd 创建并返回一个RootCmd实例，应用传入的选项。
func NewRootCmd(processName string, opts ...func(*CmdOpts)) *RootCmd {
	rootCmd := &RootCmd{processName: processName}
	cmd := cobra.Command{
		Use:  "Start openIM application",//命令名称
		Long: fmt.Sprintf(`Start %s `, processName),//命令的详细描述
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd.persistentPreRun(cmd, opts...)},//执行前运行的函数,日志与配置文件初始化。
		SilenceUsage:  true,//是否在执行时静默
		SilenceErrors: false,//是否在执行时静默错误
	}
	cmd.Flags().StringP(FlagConf, "c", "", "path of config directory")
	cmd.Flags().IntP(FlagTransferIndex, "i", 0, "process startup sequence number")

	rootCmd.Command = cmd
	return rootCmd
}

// persistentPreRun 在cobra命令执行前执行的持久化预运行逻辑，包括初始化配置和日志。
func (r *RootCmd) persistentPreRun(cmd *cobra.Command, opts ...func(*CmdOpts)) error {
	cmdOpts := r.applyOptions(opts...)
	if err := r.initializeConfiguration(cmd, cmdOpts); err != nil {
		return err
	}

	if err := r.initializeLogger(cmdOpts); err != nil {
		return errs.WrapMsg(err, "failed to initialize logger")
	}

	return nil
}

// initializeConfiguration 加载配置文件。
func (r *RootCmd) initializeConfiguration(cmd *cobra.Command, opts *CmdOpts) error {
	configDirectory, _, err := r.getFlag(cmd)
	if err != nil {
		return err
	}
	// 加载配置项
	for configFileName, configStruct := range opts.configMap {
		err := config.LoadConfig(filepath.Join(configDirectory, configFileName),
			ConfigEnvPrefixMap[configFileName], configStruct)
		if err != nil {
			return err
		}
	}
	// 加载日志配置
	return config.LoadConfig(filepath.Join(configDirectory, LogConfigFileName),
		ConfigEnvPrefixMap[LogConfigFileName], &r.log)
}

// applyOptions 应用命令行选项并返回配置。
func (r *RootCmd) applyOptions(opts ...func(*CmdOpts)) *CmdOpts {
	cmdOpts := defaultCmdOpts()
	for _, opt := range opts {
		opt(cmdOpts)
	}

	return cmdOpts
}

// initializeLogger 初始化日志系统。
func (r *RootCmd) initializeLogger(cmdOpts *CmdOpts) error {
	err := log.InitFromConfig(
		cmdOpts.loggerPrefixName,
		r.processName,
		r.log.RemainLogLevel,
		r.log.IsStdout,
		r.log.IsJson,
		r.log.StorageLocation,
		r.log.RemainRotationCount,
		r.log.RotationTime,
		config.Version,
	)
	if err != nil {
		return errs.Wrap(err)
	}
	return errs.Wrap(log.InitConsoleLogger(r.processName, r.log.RemainLogLevel, r.log.IsJson, config.Version))
}

// defaultCmdOpts 返回CmdOpts的默认配置-日志前缀。
func defaultCmdOpts() *CmdOpts {
	return &CmdOpts{
		loggerPrefixName: "openim-service-log",
	}
}

// getFlag 从cobra命令中获取配置目录路径和启动序列号。
func (r *RootCmd) getFlag(cmd *cobra.Command) (string, int, error) {
	configDirectory, err := cmd.Flags().GetString(FlagConf)
	if err != nil {
		return "", 0, errs.Wrap(err)
	}
	index, err := cmd.Flags().GetInt(FlagTransferIndex)
	if err != nil {
		return "", 0, errs.Wrap(err)
	}
	r.index = index
	return configDirectory, index, nil
}

// Execute 执行RootCmd代表的命令。
func (r *RootCmd) Execute() error {
	return r.Command.Execute()
}

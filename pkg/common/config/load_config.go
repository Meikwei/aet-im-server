package config

import (
	"strings"

	"github.com/Meikwei/go-tools/errs"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// LoadConfig 从指定路径加载配置文件，并使用给定的环境变量前缀进行环境变量配置的解析。
// 它将配置文件的内容解析到提供的 config 参数中。
// 参数:
// - path: 配置文件的路径。
// - envPrefix: 用于加载环境变量的前缀。
// - config: 指向要解析配置的目标结构体的指针。
// 返回值:
// - error: 加载或解析配置时遇到的任何错误。
func LoadConfig(path string, envPrefix string, config any) error {
	v := viper.New() // 初始化 viper 配置管理器

	v.SetConfigFile(path) // 设置配置文件路径
	v.SetEnvPrefix(envPrefix) // 设置环境变量前缀
	v.AutomaticEnv() // 自动加载环境变量
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 设置环境变量的键名替换规则，将点替换为下划线

	// 尝试读取并解析配置文件
	if err := v.ReadInConfig(); err != nil {
		return errs.WrapMsg(err, "failed to read config file", "path", path, "envPrefix", envPrefix)
	}

	// 将配置文件内容反序列化到提供的 config 参数中
	if err := v.Unmarshal(config, func(config *mapstructure.DecoderConfig) {
		config.TagName = "mapstructure" // 设置结构体标签名称
	}); err != nil {
		return errs.WrapMsg(err, "failed to unmarshal config", "path", path, "envPrefix", envPrefix)
	}
	return nil
}

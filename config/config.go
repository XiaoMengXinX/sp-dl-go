package config

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"os"
)

const ppToken = "015c8801577620d7d46de9d696bb9574"

var CM = NewConfigManager()

func init() {
	CM.Initialize()
}

// Data 定义配置数据的结构
type Data struct {
	SpDc              string `json:"sp_dc"`
	AccessToken       string `json:"accessToken"`
	AccessTokenExpire int64  `json:"accessTokenExpire"`
	PlayPlayToken     string `json:"playPlayToken"`
}

// Manager 管理配置的结构体
type Manager struct {
	configFileName string
	config         Data
	defaults       Data
}

// NewConfigManager 创建一个新的 Manager 实例，并初始化默认配置
func NewConfigManager() *Manager {
	defaults := Data{
		SpDc:              "",
		AccessToken:       "",
		AccessTokenExpire: -1,
		PlayPlayToken:     ppToken,
	}
	return &Manager{
		configFileName: "config.json",
		config:         defaults,
		defaults:       defaults,
	}
}

// Initialize 初始化配置文件，如果文件不存在则创建一个默认配置文件
func (cm *Manager) Initialize() {
	if _, err := os.Stat(cm.configFileName); errors.Is(err, os.ErrNotExist) {
		cm.writeConfig()
	}
}

// ReadConfig 从文件读取配置并解析 JSON 内容
func (cm *Manager) ReadConfig() error {
	data, err := os.ReadFile(cm.configFileName)
	if err != nil {
		return fmt.Errorf("unable to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &cm.config); err != nil {
		return fmt.Errorf("unable to parse json config file: %w", err)
	}
	return nil
}

// Get 获取当前配置的副本
func (cm *Manager) Get() Data {
	return cm.config
}

// ReadAndGet 读取配置文件并返回配置数据
func (cm *Manager) ReadAndGet() (Data, error) {
	if err := cm.ReadConfig(); err != nil {
		return Data{}, err
	}
	return cm.Get(), nil
}

// GetDefault 获取默认配置的副本
func (cm *Manager) GetDefault() Data {
	return cm.defaults
}

// writeConfig 将当前配置写入文件
func (cm *Manager) writeConfig() {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal config to json: %v", err)
	}

	if err := os.WriteFile(cm.configFileName, data, 0644); err != nil {
		log.Fatalf("Unable to write config to file: %v", err)
	}
}

// Set 更新配置中的某个值，并立即写入文件
func (cm *Manager) Set(newConfig Data) {
	cm.config = newConfig
	cm.writeConfig()
}

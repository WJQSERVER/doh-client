package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr    string `yaml:"listen_addr"`
	DnsAddr       string `yaml:"dns_addr"`
	DohServer     string `yaml:"doh_server"`
	DohServerHost string `yaml:"doh_server_host"`
	DohServerIP   string `yaml:"doh_server_ip"`
	LogFilePath   string `yaml:"log_path"`
	MaxLogSize    int64  `yaml:"max_log_size"`
}

// LoadConfig 从 YAML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	var config Config
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

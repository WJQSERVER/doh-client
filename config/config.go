package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ListenAddr  string `yaml:"listen_addr"`
	DnsAddr     string `yaml:"dns_addr"`
	DohServer   string `yaml:"doh_server"`
	LogFilePath string `yaml:"log_path"`
	MaxLogSize  int64  `yaml:"max_log_size"`
}

// LoadConfig 从 YAML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Nacos  NacosConfig  `yaml:"nacos"`
}

type NacosConfig struct {
	Enabled     bool    `yaml:"enabled"`
	ServerAddr  string  `yaml:"server-addr"`
	ServerPort  uint64  `yaml:"server-port"`
	NamespaceID string  `yaml:"namespace-id"`
	ServiceName string  `yaml:"service-name"`
	GroupName   string  `yaml:"group-name"`
	ClusterName string  `yaml:"cluster-name"`
	Weight      float64 `yaml:"weight"`
	Username    string  `yaml:"username"`
	Password    string  `yaml:"password"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func (m MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&timeout=5s&readTimeout=10s&writeTimeout=10s",
		m.User, m.Password, m.Host, m.Port, m.Database)
}

func Default() Config {
	return Config{
		Server: ServerConfig{Addr: ":9092"},
		MySQL:  MySQLConfig{Host: "127.0.0.1", Port: 3306, User: "root", Database: "dextea"},
	}
}

func Load(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
	}
	cfg := Default()
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("解析配置文件 %s 失败: %w", path, err)
	}
	if cfg.Nacos.Enabled {
		if cfg.Nacos.ServerAddr == "" {
			return Config{}, fmt.Errorf("nacos.enabled=true 时必须配置 nacos.server-addr")
		}
		if cfg.Nacos.ServerPort == 0 {
			cfg.Nacos.ServerPort = 8848
		}
		if cfg.Nacos.ServiceName == "" {
			cfg.Nacos.ServiceName = "dextea-store-service"
		}
		if cfg.Nacos.GroupName == "" {
			cfg.Nacos.GroupName = "DEFAULT_GROUP"
		}
		if cfg.Nacos.ClusterName == "" {
			cfg.Nacos.ClusterName = "DEFAULT"
		}
		if cfg.Nacos.Weight <= 0 {
			cfg.Nacos.Weight = 1
		}
	}
	return cfg, nil
}

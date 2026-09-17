package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	MySQL  MySQLConfig  `yaml:"mysql"`
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
	return cfg, nil
}

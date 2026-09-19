package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Nacos  NacosConfig  `yaml:"nacos"`
	Auth   AuthConfig   `yaml:"auth"`
}

type AuthConfig struct {
	Enabled         bool   `yaml:"enabled"`
	AdminToken      string `yaml:"admin-token"`
	BusinessToken   string `yaml:"business-token"`
	CredentialToken string `yaml:"credential-token"`
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
	InstanceIP  string  `yaml:"instance-ip"`
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
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("加载 .env 失败: %w", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
	}
	cfg := Default()
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("解析配置文件 %s 失败: %w", path, err)
	}
	if err := applyNacosEnv(&cfg.Nacos); err != nil {
		return Config{}, err
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
	cfg.Auth.AdminToken = strings.TrimSpace(cfg.Auth.AdminToken)
	cfg.Auth.BusinessToken = strings.TrimSpace(cfg.Auth.BusinessToken)
	cfg.Auth.CredentialToken = strings.TrimSpace(cfg.Auth.CredentialToken)
	if cfg.Auth.Enabled && (cfg.Auth.AdminToken == "" || cfg.Auth.BusinessToken == "" || cfg.Auth.CredentialToken == "") {
		return Config{}, fmt.Errorf("auth.enabled=true 时必须分别配置 admin-token、business-token 和 credential-token")
	}
	return cfg, nil
}

// applyNacosEnv 统一读取 Nacos 环境变量。godotenv.Load 不会覆盖启动进程已有的
// 系统环境变量，因此系统环境变量优先于 .env。
func applyNacosEnv(cfg *NacosConfig) error {
	if value, ok := os.LookupEnv("NACOS_ENABLED"); ok && strings.TrimSpace(value) != "" {
		enabled, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("NACOS_ENABLED 必须是 true 或 false: %w", err)
		}
		cfg.Enabled = enabled
	}
	if value, ok := os.LookupEnv("NACOS_SERVER_ADDR"); ok {
		host, port, err := net.SplitHostPort(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("NACOS_SERVER_ADDR 必须是 host:port，例如 127.0.0.1:8848: %w", err)
		}
		serverPort, err := strconv.ParseUint(port, 10, 16)
		if err != nil || serverPort == 0 {
			return fmt.Errorf("NACOS_SERVER_ADDR 端口无效: %q", port)
		}
		cfg.ServerAddr = host
		cfg.ServerPort = serverPort
	}
	if value, ok := os.LookupEnv("NACOS_NAMESPACE"); ok {
		cfg.NamespaceID = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("NACOS_GROUP"); ok {
		cfg.GroupName = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("NACOS_CLUSTER"); ok {
		cfg.ClusterName = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("NACOS_SERVICE_NAME"); ok {
		cfg.ServiceName = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("NACOS_USERNAME"); ok {
		cfg.Username = value
	}
	if value, ok := os.LookupEnv("NACOS_PASSWORD"); ok {
		cfg.Password = value
	}
	if value, ok := os.LookupEnv("NACOS_INSTANCE_IP"); ok {
		cfg.InstanceIP = strings.TrimSpace(value)
	}
	return nil
}

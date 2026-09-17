package config

// Config 是服务配置的组合根。随着基础设施接入，在这里扩展数据库、Redis、Nacos 等配置。
type Config struct {
	Server ServerConfig
}

type ServerConfig struct {
	Addr string
}

func Default() Config {
	return Config{
		Server: ServerConfig{Addr: ":9092"},
	}
}

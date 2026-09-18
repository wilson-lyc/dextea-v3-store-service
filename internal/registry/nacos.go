package registry

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/wilson-lyc/dextea-store-service/internal/config"
)

type Registrar struct {
	client naming_client.INamingClient
	cfg    config.NacosConfig
	ip     string
	port   uint64
}

func Register(cfg config.NacosConfig, listenAddr string) (*Registrar, error) {
	ip, port, err := endpoint(listenAddr)
	if err != nil {
		return nil, err
	}
	clientCfg := constant.NewClientConfig(constant.WithNamespaceId(cfg.NamespaceID), constant.WithNotLoadCacheAtStart(true), constant.WithUsername(cfg.Username), constant.WithPassword(cfg.Password))
	serverCfg := constant.NewServerConfig(cfg.ServerAddr, cfg.ServerPort)
	client, err := clients.NewNamingClient(vo.NacosClientParam{ClientConfig: clientCfg, ServerConfigs: []constant.ServerConfig{*serverCfg}})
	if err != nil {
		return nil, fmt.Errorf("创建 nacos 客户端失败: %w", err)
	}
	ok, err := client.RegisterInstance(vo.RegisterInstanceParam{Ip: ip, Port: port, ServiceName: cfg.ServiceName, GroupName: cfg.GroupName, ClusterName: cfg.ClusterName, Weight: cfg.Weight, Enable: true, Healthy: true, Ephemeral: true, Metadata: map[string]string{"protocol": "grpc"}})
	if err != nil || !ok {
		return nil, fmt.Errorf("注册到 nacos 失败: err=%v ok=%v", err, ok)
	}
	log.Printf("[info] registered %s (%s:%d) to nacos %s:%d/%s", cfg.ServiceName, ip, port, cfg.ServerAddr, cfg.ServerPort, cfg.GroupName)
	return &Registrar{client: client, cfg: cfg, ip: ip, port: port}, nil
}

func (r *Registrar) Deregister() {
	ok, err := r.client.DeregisterInstance(vo.DeregisterInstanceParam{Ip: r.ip, Port: r.port, ServiceName: r.cfg.ServiceName, GroupName: r.cfg.GroupName, Cluster: r.cfg.ClusterName, Ephemeral: true})
	if err != nil || !ok {
		log.Printf("[error] deregister from nacos: err=%v ok=%v", err, ok)
	}
}

func endpoint(addr string) (string, uint64, error) {
	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("解析监听地址 %s 失败: %w", addr, err)
	}
	port, err := strconv.ParseUint(portText, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("解析监听端口 %s 失败: %w", portText, err)
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			return "", 0, fmt.Errorf("获取本机 IP 失败: %w", err)
		}
		host = conn.LocalAddr().(*net.UDPAddr).IP.String()
		_ = conn.Close()
	}
	return host, port, nil
}

// Package rpc contains the transport adapters for the store domain.
package rpc

// Server 是门店领域 gRPC 服务的组合根预留位。
// 具体 protobuf 生成的服务接口确认后，在这里实现对应的 RPC 方法。
type Server struct{}

func NewServer() *Server {
	return &Server{}
}

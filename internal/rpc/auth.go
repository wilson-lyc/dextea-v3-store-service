package rpc

import (
	"context"
	"crypto/subtle"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const serviceTokenHeader = "x-service-token"

type RoleTokens struct {
	Admin      string
	Business   string
	Credential string
}

// RoleTokenInterceptor applies different service tokens to the management,
// business-read, and credential planes.
func RoleTokenInterceptor(tokens RoleTokens) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		expected := ""
		switch {
		case strings.HasPrefix(info.FullMethod, "/dextea.store.v1.StoreAdminService/"):
			expected = tokens.Admin
		case strings.HasPrefix(info.FullMethod, "/dextea.store.v1.StoreBusinessService/"):
			expected = tokens.Business
		case strings.HasPrefix(info.FullMethod, "/dextea.store.v1.StoreCredentialService/"):
			expected = tokens.Credential
		}
		if expected == "" {
			return nil, status.Error(codes.Unauthenticated, "服务认证令牌未配置")
		}

		incoming, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(incoming.Get(serviceTokenHeader)) == 0 {
			return nil, status.Error(codes.Unauthenticated, "缺少服务认证令牌")
		}
		provided := incoming.Get(serviceTokenHeader)[0]
		if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			return nil, status.Error(codes.Unauthenticated, "服务认证令牌无效")
		}
		return handler(ctx, req)
	}
}

func isPublicMethod(method string) bool {
	return strings.HasPrefix(method, "/grpc.health.v1.Health/") ||
		strings.HasPrefix(method, "/grpc.reflection.v1alpha.ServerReflection/")
}

package rpc

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRoleTokenInterceptorRequiresToken(t *testing.T) {
	interceptor := RoleTokenInterceptor(RoleTokens{Admin: "secret", Business: "secret", Credential: "secret"})
	handler := func(context.Context, any) (any, error) { return "ok", nil }
	info := &grpc.UnaryServerInfo{FullMethod: "/dextea.store.v1.StoreAdminService/GetStore"}

	if _, err := interceptor(context.Background(), nil, info, handler); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("missing token code = %v, want %v", status.Code(err), codes.Unauthenticated)
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(serviceTokenHeader, "secret"))
	if value, err := interceptor(ctx, nil, info, handler); err != nil || value != "ok" {
		t.Fatalf("valid token result = %v, %v", value, err)
	}
}

func TestRoleTokenInterceptorUsesPlaneSpecificTokens(t *testing.T) {
	interceptor := RoleTokenInterceptor(RoleTokens{
		Admin:      "admin-secret",
		Business:   "business-secret",
		Credential: "credential-secret",
	})
	handler := func(context.Context, any) (any, error) { return "ok", nil }

	adminInfo := &grpc.UnaryServerInfo{FullMethod: "/dextea.store.v1.StoreAdminService/GetStore"}
	wrong := metadata.NewIncomingContext(context.Background(), metadata.Pairs(serviceTokenHeader, "business-secret"))
	if _, err := interceptor(wrong, nil, adminInfo, handler); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("wrong admin token code = %v, want %v", status.Code(err), codes.Unauthenticated)
	}
	admin := metadata.NewIncomingContext(context.Background(), metadata.Pairs(serviceTokenHeader, "admin-secret"))
	if value, err := interceptor(admin, nil, adminInfo, handler); err != nil || value != "ok" {
		t.Fatalf("admin token result = %v, %v", value, err)
	}

	businessInfo := &grpc.UnaryServerInfo{FullMethod: "/dextea.store.v1.StoreBusinessService/SearchStores"}
	business := metadata.NewIncomingContext(context.Background(), metadata.Pairs(serviceTokenHeader, "business-secret"))
	if value, err := interceptor(business, nil, businessInfo, handler); err != nil || value != "ok" {
		t.Fatalf("business token result = %v, %v", value, err)
	}
}

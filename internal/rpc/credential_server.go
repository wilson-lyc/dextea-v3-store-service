package rpc

import (
	"context"

	storev1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/store/v1"

	"github.com/wilson-lyc/dextea-store-service/internal/service"
)

// CredentialServer isolates store-account authentication and self-service
// password changes from both admin management and customer reads.
type CredentialServer struct {
	storev1.UnimplementedStoreCredentialServiceServer
	service *service.StoreService
}

func NewCredentialServer(svc *service.StoreService) *CredentialServer {
	return &CredentialServer{service: svc}
}

func (s *CredentialServer) AuthenticateStore(ctx context.Context, in *storev1.AuthenticateStoreRequest) (*storev1.StoreAuthInfo, error) {
	store, err := s.service.Authenticate(ctx, in.Account, in.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &storev1.StoreAuthInfo{StoreId: store.ID, Status: store.Status, Name: store.Name}, nil
}

func (s *CredentialServer) ChangeStorePassword(ctx context.Context, in *storev1.ChangeStorePasswordRequest) (*storev1.PasswordChangedResponse, error) {
	if _, err := s.service.ChangePassword(ctx, in.Id, in.OldPassword, in.NewPassword); err != nil {
		return nil, mapError(err)
	}
	return &storev1.PasswordChangedResponse{Changed: true}, nil
}

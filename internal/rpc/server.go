// Package rpc contains the gRPC transport adapter for the store domain.
package rpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	storev1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/store/v1"

	"github.com/wilson-lyc/dextea-store-service/internal/model"
	"github.com/wilson-lyc/dextea-store-service/internal/service"
)

type Server struct {
	storev1.UnimplementedStoreServiceServer
	service *service.StoreService
}

func NewServer(svc *service.StoreService) *Server { return &Server{service: svc} }

func (s *Server) CreateStore(ctx context.Context, in *storev1.CreateStoreRequest) (*storev1.CreateStoreResponse, error) {
	statusValue := service.StatusPending
	if in.Status != nil {
		statusValue = in.GetStatus()
	}
	store, password, err := s.service.Create(ctx, model.Store{
		Name: in.Name, Province: in.Province, City: in.City, District: in.District,
		Address: in.Address, Status: statusValue, BusinessHours: in.BusinessHours,
		Phone: in.Phone, Longitude: in.Longitude, Latitude: in.Latitude,
		Account: in.Account, Email: in.Email,
	}, in.InitialPassword)
	if err != nil {
		return nil, mapError(err)
	}
	return &storev1.CreateStoreResponse{Store: view(store), InitialPassword: password}, nil
}

func (s *Server) GetStore(ctx context.Context, in *storev1.GetStoreRequest) (*storev1.Store, error) {
	store, err := s.service.Get(ctx, in.Id)
	if err != nil {
		return nil, mapError(err)
	}
	return view(store), nil
}

func (s *Server) GetStoreByAccount(ctx context.Context, in *storev1.GetStoreByAccountRequest) (*storev1.Store, error) {
	store, err := s.service.GetByAccount(ctx, in.Account)
	if err != nil {
		return nil, mapError(err)
	}
	return view(store), nil
}

func (s *Server) ListStores(ctx context.Context, in *storev1.ListStoresRequest) (*storev1.ListStoresResponse, error) {
	stores, total, err := s.service.List(ctx, in.Page, in.PageSize, in.Keyword, in.Status)
	if err != nil {
		return nil, mapError(err)
	}
	page, pageSize := in.Page, in.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	out := &storev1.ListStoresResponse{Total: total, Page: page, PageSize: pageSize}
	for _, item := range stores {
		out.Stores = append(out.Stores, view(&item))
	}
	return out, nil
}

func (s *Server) SearchStores(ctx context.Context, in *storev1.SearchStoresRequest) (*storev1.SearchStoresResponse, error) {
	stores, err := s.service.Search(ctx, in.City, in.Keyword, in.IncludeUnavailable)
	if err != nil {
		return nil, mapError(err)
	}
	out := &storev1.SearchStoresResponse{}
	for _, item := range stores {
		out.Stores = append(out.Stores, view(&item))
	}
	return out, nil
}

func (s *Server) GetNearbyStores(ctx context.Context, in *storev1.GetNearbyStoresRequest) (*storev1.GetNearbyStoresResponse, error) {
	stores, distances, err := s.service.Nearby(ctx, in.Longitude, in.Latitude, in.DistanceKm, in.Count)
	if err != nil {
		return nil, mapError(err)
	}
	out := &storev1.GetNearbyStoresResponse{}
	for i, item := range stores {
		out.Stores = append(out.Stores, &storev1.NearbyStore{Store: view(&item), DistanceKm: distances[i]})
	}
	return out, nil
}

func (s *Server) UpdateStoreProfile(ctx context.Context, in *storev1.UpdateStoreProfileRequest) (*storev1.Store, error) {
	fields := map[string]any{}
	if in.Name != nil {
		fields["name"] = in.GetName()
	}
	if in.Phone != nil {
		fields["phone"] = in.GetPhone()
	}
	if in.BusinessHours != nil {
		fields["business_hours"] = in.GetBusinessHours()
	}
	if in.Email != nil {
		fields["email"] = in.GetEmail()
	}
	store, err := s.service.UpdateProfile(ctx, in.Id, fields)
	if err != nil {
		return nil, mapError(err)
	}
	return view(store), nil
}

func (s *Server) UpdateStoreLocation(ctx context.Context, in *storev1.UpdateStoreLocationRequest) (*storev1.Store, error) {
	fields := map[string]any{"longitude": in.Longitude, "latitude": in.Latitude}
	if in.Province != nil {
		fields["province"] = in.GetProvince()
	}
	if in.City != nil {
		fields["city"] = in.GetCity()
	}
	if in.District != nil {
		fields["district"] = in.GetDistrict()
	}
	if in.Address != nil {
		fields["address"] = in.GetAddress()
	}
	store, err := s.service.UpdateLocation(ctx, in.Id, fields)
	if err != nil {
		return nil, mapError(err)
	}
	return view(store), nil
}

func (s *Server) UpdateStoreStatus(ctx context.Context, in *storev1.UpdateStoreStatusRequest) (*storev1.Store, error) {
	store, err := s.service.UpdateStatus(ctx, in.Id, in.Status)
	if err != nil {
		return nil, mapError(err)
	}
	return view(store), nil
}

func (s *Server) ResetStorePassword(ctx context.Context, in *storev1.ResetStorePasswordRequest) (*storev1.ResetStorePasswordResponse, error) {
	_, password, err := s.service.ResetPassword(ctx, in.Id, in.OldPassword, in.GetNewPassword())
	if err != nil {
		return nil, mapError(err)
	}
	return &storev1.ResetStorePasswordResponse{Password: password}, nil
}

func view(store *model.Store) *storev1.Store {
	if store == nil {
		return nil
	}
	return &storev1.Store{
		Id: store.ID, Name: store.Name, Province: store.Province, City: store.City,
		District: store.District, Address: store.Address, Status: store.Status,
		BusinessHours: store.BusinessHours, Phone: store.Phone, Longitude: store.Longitude,
		Latitude: store.Latitude, Account: store.Account, Email: store.Email,
		CreatedAt: store.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: store.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalid):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrPassword):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "门店服务内部错误")
	}
}

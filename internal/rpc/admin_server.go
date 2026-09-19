package rpc

import (
	"context"

	storev1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/store/v1"

	"github.com/wilson-lyc/dextea-store-service/internal/model"
	"github.com/wilson-lyc/dextea-store-service/internal/repository"
	"github.com/wilson-lyc/dextea-store-service/internal/service"
)

// AdminServer implements the management plane of the store aggregate.
type AdminServer struct {
	storev1.UnimplementedStoreAdminServiceServer
	service *service.StoreService
}

func NewAdminServer(svc *service.StoreService) *AdminServer {
	return &AdminServer{service: svc}
}

func (s *AdminServer) CreateStore(ctx context.Context, in *storev1.CreateStoreRequest) (*storev1.CreateStoreResponse, error) {
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
	return &storev1.CreateStoreResponse{Store: adminView(store), InitialPassword: password}, nil
}

func (s *AdminServer) GetStore(ctx context.Context, in *storev1.GetStoreRequest) (*storev1.Store, error) {
	store, err := s.service.Get(ctx, in.Id)
	if err != nil {
		return nil, mapError(err)
	}
	return adminView(store), nil
}

func (s *AdminServer) GetStoreByAccount(ctx context.Context, in *storev1.GetStoreByAccountRequest) (*storev1.Store, error) {
	store, err := s.service.GetByAccount(ctx, in.Account)
	if err != nil {
		return nil, mapError(err)
	}
	return adminView(store), nil
}

func (s *AdminServer) GetStores(ctx context.Context, in *storev1.GetAdminStoresRequest) (*storev1.GetAdminStoresResponse, error) {
	stores, err := s.service.GetManyAdmin(ctx, in.Ids, in.IncludeUnavailable)
	if err != nil {
		return nil, mapError(err)
	}
	out := &storev1.GetAdminStoresResponse{Stores: make([]*storev1.Store, 0, len(stores))}
	for i := range stores {
		out.Stores = append(out.Stores, adminView(&stores[i]))
	}
	return out, nil
}

func (s *AdminServer) SearchStores(ctx context.Context, in *storev1.SearchAdminStoresRequest) (*storev1.SearchAdminStoresResponse, error) {
	filter := repository.StoreAdminSearchFilter{
		Province:           in.Province,
		City:               in.City,
		District:           in.District,
		Keyword:            in.Keyword,
		IncludeUnavailable: in.IncludeUnavailable,
	}
	if in.Status != nil {
		statusValue := in.GetStatus()
		filter.Status = &statusValue
	}
	stores, err := s.service.SearchAdmin(ctx, filter)
	if err != nil {
		return nil, mapError(err)
	}
	out := &storev1.SearchAdminStoresResponse{Stores: make([]*storev1.Store, 0, len(stores))}
	for i := range stores {
		out.Stores = append(out.Stores, adminView(&stores[i]))
	}
	return out, nil
}

func (s *AdminServer) ListStores(ctx context.Context, in *storev1.ListStoresRequest) (*storev1.ListStoresResponse, error) {
	filter := repository.StoreListFilter{Keyword: in.Keyword}
	if in.Status != nil {
		statusValue := in.GetStatus()
		filter.Status = &statusValue
	}
	if in.Province != nil {
		filter.Province = in.GetProvince()
	}
	if in.City != nil {
		filter.City = in.GetCity()
	}
	if in.District != nil {
		filter.District = in.GetDistrict()
	}
	stores, total, err := s.service.List(ctx, in.Page, in.PageSize, filter)
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
	for i := range stores {
		out.Stores = append(out.Stores, adminView(&stores[i]))
	}
	return out, nil
}

func (s *AdminServer) GetStoreStatistics(ctx context.Context, in *storev1.GetStoreStatisticsRequest) (*storev1.GetStoreStatisticsResponse, error) {
	_ = in
	counts, err := s.service.Statistics(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	out := &storev1.GetStoreStatisticsResponse{Statuses: make([]*storev1.StoreStatusCount, 0, len(counts))}
	for _, count := range counts {
		out.Statuses = append(out.Statuses, &storev1.StoreStatusCount{Status: count.Status, Count: count.Count})
	}
	return out, nil
}

func (s *AdminServer) UpdateStoreProfile(ctx context.Context, in *storev1.UpdateStoreProfileRequest) (*storev1.Store, error) {
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
	return adminView(store), nil
}

func (s *AdminServer) UpdateStoreLocation(ctx context.Context, in *storev1.UpdateStoreLocationRequest) (*storev1.Store, error) {
	fields := map[string]any{}
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
	if in.Longitude != nil {
		fields["longitude"] = in.GetLongitude()
	}
	if in.Latitude != nil {
		fields["latitude"] = in.GetLatitude()
	}
	store, err := s.service.UpdateLocation(ctx, in.Id, fields)
	if err != nil {
		return nil, mapError(err)
	}
	return adminView(store), nil
}

func (s *AdminServer) UpdateStoreStatus(ctx context.Context, in *storev1.UpdateStoreStatusRequest) (*storev1.Store, error) {
	store, err := s.service.UpdateStatus(ctx, in.Id, in.Status)
	if err != nil {
		return nil, mapError(err)
	}
	return adminView(store), nil
}

func (s *AdminServer) ResetStorePassword(ctx context.Context, in *storev1.ResetStorePasswordRequest) (*storev1.ResetStorePasswordResponse, error) {
	_, password, err := s.service.ResetPassword(ctx, in.Id, "", in.GetNewPassword())
	if err != nil {
		return nil, mapError(err)
	}
	return &storev1.ResetStorePasswordResponse{Password: password}, nil
}

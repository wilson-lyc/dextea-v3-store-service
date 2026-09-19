package rpc

import (
	"context"

	storev1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/store/v1"

	"github.com/wilson-lyc/dextea-store-service/internal/repository"
	"github.com/wilson-lyc/dextea-store-service/internal/service"
)

// BusinessServer implements the customer-facing, read-only store plane.
type BusinessServer struct {
	storev1.UnimplementedStoreBusinessServiceServer
	service *service.StoreService
}

func NewBusinessServer(svc *service.StoreService) *BusinessServer {
	return &BusinessServer{service: svc}
}

func (s *BusinessServer) GetStore(ctx context.Context, in *storev1.GetBusinessStoreRequest) (*storev1.BusinessStoreDistance, error) {
	store, distance, err := s.service.GetWithDistance(ctx, in.Id, in.Longitude, in.Latitude)
	if err != nil {
		return nil, mapError(err)
	}
	return &storev1.BusinessStoreDistance{Store: businessView(store), DistanceKm: distance}, nil
}

func (s *BusinessServer) GetStores(ctx context.Context, in *storev1.GetBusinessStoresRequest) (*storev1.GetBusinessStoresResponse, error) {
	stores, err := s.service.GetMany(ctx, in.Ids)
	if err != nil {
		return nil, mapError(err)
	}
	out := &storev1.GetBusinessStoresResponse{Stores: make([]*storev1.BusinessStore, 0, len(stores))}
	for i := range stores {
		out.Stores = append(out.Stores, businessView(&stores[i]))
	}
	return out, nil
}

func (s *BusinessServer) SearchStores(ctx context.Context, in *storev1.SearchBusinessStoresRequest) (*storev1.SearchBusinessStoresResponse, error) {
	filter := repository.StoreBusinessSearchFilter{
		Province: in.Province,
		City:     in.City,
		District: in.District,
		Keyword:  in.Keyword,
	}
	stores, distances, err := s.service.SearchWithDistance(ctx, in.Longitude, in.Latitude, filter)
	if err != nil {
		return nil, mapError(err)
	}
	out := &storev1.SearchBusinessStoresResponse{Stores: make([]*storev1.BusinessStoreDistance, 0, len(stores))}
	for i := range stores {
		out.Stores = append(out.Stores, &storev1.BusinessStoreDistance{
			Store:      businessView(&stores[i]),
			DistanceKm: distances[i],
		})
	}
	return out, nil
}

func (s *BusinessServer) ListStoreCities(ctx context.Context, in *storev1.ListStoreCitiesRequest) (*storev1.ListStoreCitiesResponse, error) {
	_ = in
	cities, err := s.service.ListCities(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	return &storev1.ListStoreCitiesResponse{Cities: cities}, nil
}

func (s *BusinessServer) GetNearbyStores(ctx context.Context, in *storev1.GetBusinessNearbyStoresRequest) (*storev1.GetBusinessNearbyStoresResponse, error) {
	stores, distances, err := s.service.Nearby(ctx, in.Longitude, in.Latitude, in.DistanceKm, in.Count)
	if err != nil {
		return nil, mapError(err)
	}
	out := &storev1.GetBusinessNearbyStoresResponse{Stores: make([]*storev1.BusinessStoreDistance, 0, len(stores))}
	for i := range stores {
		out.Stores = append(out.Stores, &storev1.BusinessStoreDistance{
			Store:      businessView(&stores[i]),
			DistanceKm: distances[i],
		})
	}
	return out, nil
}

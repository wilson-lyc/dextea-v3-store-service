package rpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	storev1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/store/v1"

	"github.com/wilson-lyc/dextea-store-service/internal/model"
	"github.com/wilson-lyc/dextea-store-service/internal/service"
)

func adminView(store *model.Store) *storev1.Store {
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

func businessView(store *model.Store) *storev1.BusinessStore {
	if store == nil {
		return nil
	}
	return &storev1.BusinessStore{
		Id: store.ID, Name: store.Name, Province: store.Province, City: store.City,
		District: store.District, Address: store.Address, Status: store.Status,
		BusinessHours: store.BusinessHours, Phone: store.Phone,
		Longitude: store.Longitude, Latitude: store.Latitude,
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

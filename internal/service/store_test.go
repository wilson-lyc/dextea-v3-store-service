package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/wilson-lyc/dextea-store-service/internal/model"
	"github.com/wilson-lyc/dextea-store-service/internal/repository"
)

type fakeStoreRepository struct {
	stores       []model.Store
	statusCounts []model.StoreStatusCount
	search       []model.Store
	updated      map[string]any
}

func (f *fakeStoreRepository) Create(context.Context, *model.Store) (uint64, error) {
	return 0, nil
}

func (f *fakeStoreRepository) FindByID(_ context.Context, id uint64) (*model.Store, error) {
	for i := range f.stores {
		if f.stores[i].ID == id {
			return &f.stores[i], nil
		}
	}
	return nil, nil
}

func (f *fakeStoreRepository) FindByAccount(context.Context, string) (*model.Store, error) {
	return nil, nil
}

func (f *fakeStoreRepository) FindByIDs(_ context.Context, ids []uint64, includeUnavailable bool) ([]model.Store, error) {
	result := make([]model.Store, 0, len(ids))
	for _, store := range f.stores {
		if !includeUnavailable && !store.Available() {
			continue
		}
		for _, id := range ids {
			if store.ID == id {
				result = append(result, store)
				break
			}
		}
	}
	return result, nil
}

func (f *fakeStoreRepository) List(context.Context, int32, int32, repository.StoreListFilter) ([]model.Store, int64, error) {
	return nil, 0, nil
}

func (f *fakeStoreRepository) Search(context.Context, repository.StoreBusinessSearchFilter) ([]model.Store, error) {
	return f.search, nil
}

func (f *fakeStoreRepository) SearchAdmin(context.Context, repository.StoreAdminSearchFilter) ([]model.Store, error) {
	return f.search, nil
}

func (f *fakeStoreRepository) ListCities(context.Context) ([]string, error) {
	return nil, nil
}

func (f *fakeStoreRepository) CountByStatus(context.Context) ([]model.StoreStatusCount, error) {
	return f.statusCounts, nil
}

func (f *fakeStoreRepository) Update(_ context.Context, id uint64, fields map[string]any) (*model.Store, error) {
	f.updated = fields
	for i := range f.stores {
		if f.stores[i].ID == id {
			return &f.stores[i], nil
		}
	}
	return nil, nil
}

func TestGetManyPreservesRequestOrderAndDeduplicates(t *testing.T) {
	repo := &fakeStoreRepository{stores: []model.Store{
		{ID: 1, Status: StatusOpen},
		{ID: 2, Status: StatusDefunct},
		{ID: 3, Status: StatusClosed},
	}}
	svc := NewStoreService(repo)

	stores, err := svc.GetMany(context.Background(), []uint64{3, 1, 3})
	if err != nil {
		t.Fatal(err)
	}
	got := []uint64{stores[0].ID, stores[1].ID}
	if want := []uint64{3, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got IDs %v, want %v", got, want)
	}
}

func TestStatisticsReturnsStableStatusSet(t *testing.T) {
	repo := &fakeStoreRepository{statusCounts: []model.StoreStatusCount{{Status: StatusOpen, Count: 4}}}
	svc := NewStoreService(repo)

	counts, err := svc.Statistics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(counts) != 4 || counts[1].Status != StatusOpen || counts[1].Count != 4 || counts[3].Count != 0 {
		t.Fatalf("unexpected statistics: %#v", counts)
	}
}

func TestUpdateLocationRequiresCoordinatePair(t *testing.T) {
	repo := &fakeStoreRepository{stores: []model.Store{{ID: 1, Status: StatusOpen}}}
	svc := NewStoreService(repo)

	if _, err := svc.UpdateLocation(context.Background(), 1, map[string]any{"longitude": 120.1}); err != ErrInvalid {
		t.Fatalf("one coordinate error = %v, want %v", err, ErrInvalid)
	}
	if _, err := svc.UpdateLocation(context.Background(), 1, map[string]any{"address": "新地址"}); err != nil {
		t.Fatalf("text-only location update failed: %v", err)
	}
}

func TestSearchWithDistanceSortsAvailableStores(t *testing.T) {
	repo := &fakeStoreRepository{search: []model.Store{
		{ID: 1, Status: StatusOpen, Longitude: 121.0, Latitude: 31.0},
		{ID: 2, Status: StatusOpen, Longitude: 120.5, Latitude: 30.5},
	}}
	svc := NewStoreService(repo)

	stores, distances, err := svc.SearchWithDistance(context.Background(), 120.5, 30.5, repository.StoreBusinessSearchFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(stores) != 2 || stores[0].ID != 2 || len(distances) != 2 || distances[0] > distances[1] {
		t.Fatalf("unexpected distance search result: ids=%v distances=%v", []uint64{stores[0].ID, stores[1].ID}, distances)
	}
}

func TestGetWithDistanceHidesUnavailableStore(t *testing.T) {
	repo := &fakeStoreRepository{stores: []model.Store{{ID: 1, Status: StatusDefunct, Longitude: 120, Latitude: 30}}}
	svc := NewStoreService(repo)

	if _, _, err := svc.GetWithDistance(context.Background(), 1, 120, 30); err != ErrNotFound {
		t.Fatalf("unavailable store error = %v, want %v", err, ErrNotFound)
	}
}

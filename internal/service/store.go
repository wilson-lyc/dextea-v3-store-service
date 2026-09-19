package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/wilson-lyc/dextea-store-service/internal/model"
	"github.com/wilson-lyc/dextea-store-service/internal/repository"
)

const (
	StatusClosed  int32 = 0
	StatusOpen    int32 = 1
	StatusPending int32 = 2
	StatusDefunct int32 = 3
)

var ErrNotFound = errors.New("门店不存在")
var ErrInvalid = errors.New("门店参数不合法")
var ErrConflict = errors.New("门店账号已存在")
var ErrPassword = errors.New("密码校验失败")

const maxBatchStoreIDs = 500

type StoreService struct{ repo repository.StoreRepository }

func NewStoreService(repo repository.StoreRepository) *StoreService {
	return &StoreService{repo: repo}
}

func (s *StoreService) Create(ctx context.Context, input model.Store, initialPassword string) (*model.Store, string, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Account = strings.TrimSpace(input.Account)
	if input.Name == "" || input.Account == "" {
		return nil, "", ErrInvalid
	}
	if !validCoordinates(input.Longitude, input.Latitude) {
		return nil, "", ErrInvalid
	}
	if !validStatus(input.Status) {
		return nil, "", ErrInvalid
	}
	if existing, err := s.repo.FindByAccount(ctx, input.Account); err != nil {
		return nil, "", err
	} else if existing != nil {
		return nil, "", ErrConflict
	}
	if initialPassword == "" {
		initialPassword = randomPassword()
	}
	input.PasswordHash = hashPassword(initialPassword)
	id, err := s.repo.Create(ctx, &input)
	if err != nil {
		return nil, "", err
	}
	created, err := s.repo.FindByID(ctx, id)
	return created, initialPassword, err
}

func (s *StoreService) Get(ctx context.Context, id uint64) (*model.Store, error) {
	if id == 0 {
		return nil, ErrInvalid
	}
	store, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, ErrNotFound
	}
	return store, nil
}

func (s *StoreService) GetByAccount(ctx context.Context, account string) (*model.Store, error) {
	if strings.TrimSpace(account) == "" {
		return nil, ErrInvalid
	}
	store, err := s.repo.FindByAccount(ctx, strings.TrimSpace(account))
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, ErrNotFound
	}
	return store, nil
}

// GetMany returns existing stores in the same order as the requested IDs.
// Missing IDs are omitted. By default only customer-visible stores are returned.
func (s *StoreService) GetMany(ctx context.Context, ids []uint64) ([]model.Store, error) {
	if len(ids) > maxBatchStoreIDs {
		return nil, ErrInvalid
	}
	uniqueIDs := make([]uint64, 0, len(ids))
	seen := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return nil, ErrInvalid
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	if len(uniqueIDs) == 0 {
		return []model.Store{}, nil
	}
	stores, err := s.repo.FindByIDs(ctx, uniqueIDs, false)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint64]model.Store, len(stores))
	for _, store := range stores {
		byID[store.ID] = store
	}
	ordered := make([]model.Store, 0, len(stores))
	for _, id := range uniqueIDs {
		if store, ok := byID[id]; ok {
			ordered = append(ordered, store)
		}
	}
	return ordered, nil
}

func (s *StoreService) GetManyAdmin(ctx context.Context, ids []uint64, includeUnavailable bool) ([]model.Store, error) {
	if len(ids) > maxBatchStoreIDs {
		return nil, ErrInvalid
	}
	uniqueIDs := make([]uint64, 0, len(ids))
	seen := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return nil, ErrInvalid
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	if len(uniqueIDs) == 0 {
		return []model.Store{}, nil
	}
	stores, err := s.repo.FindByIDs(ctx, uniqueIDs, includeUnavailable)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint64]model.Store, len(stores))
	for _, store := range stores {
		byID[store.ID] = store
	}
	ordered := make([]model.Store, 0, len(stores))
	for _, id := range uniqueIDs {
		if store, ok := byID[id]; ok {
			ordered = append(ordered, store)
		}
	}
	return ordered, nil
}

// Authenticate 校验门店凭证。Token 签发仍由调用方的 BFF 负责。
func (s *StoreService) Authenticate(ctx context.Context, account, password string) (*model.Store, error) {
	store, err := s.GetByAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	if !store.Available() || !verifyPassword(password, store.PasswordHash) {
		return nil, ErrPassword
	}
	return store, nil
}

func (s *StoreService) List(ctx context.Context, page, pageSize int32, filter repository.StoreListFilter) ([]model.Store, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if filter.Status != nil && !validStatus(*filter.Status) {
		return nil, 0, ErrInvalid
	}
	return s.repo.List(ctx, page, pageSize, filter)
}

func (s *StoreService) Search(ctx context.Context, filter repository.StoreBusinessSearchFilter) ([]model.Store, error) {
	return s.repo.Search(ctx, filter)
}

func (s *StoreService) SearchAdmin(ctx context.Context, filter repository.StoreAdminSearchFilter) ([]model.Store, error) {
	if filter.Status != nil && !validStatus(*filter.Status) {
		return nil, ErrInvalid
	}
	return s.repo.SearchAdmin(ctx, filter)
}

func (s *StoreService) ListCities(ctx context.Context) ([]string, error) {
	return s.repo.ListCities(ctx)
}

func (s *StoreService) Statistics(ctx context.Context) ([]model.StoreStatusCount, error) {
	counts, err := s.repo.CountByStatus(ctx)
	if err != nil {
		return nil, err
	}
	byStatus := make(map[int32]int64, len(counts))
	for _, count := range counts {
		byStatus[count.Status] = count.Count
	}
	statuses := []int32{StatusClosed, StatusOpen, StatusPending, StatusDefunct}
	result := make([]model.StoreStatusCount, 0, len(statuses))
	for _, status := range statuses {
		result = append(result, model.StoreStatusCount{Status: status, Count: byStatus[status]})
	}
	return result, nil
}

func (s *StoreService) Nearby(ctx context.Context, longitude, latitude, distanceKm float64, count int32) ([]model.Store, []float64, error) {
	if !validCoordinate(longitude, latitude) || distanceKm <= 0 || distanceKm > 1000 || count <= 0 || count > 100 {
		return nil, nil, ErrInvalid
	}
	stores, err := s.repo.Search(ctx, repository.StoreBusinessSearchFilter{})
	if err != nil {
		return nil, nil, err
	}
	type candidate struct {
		store    model.Store
		distance float64
	}
	candidates := make([]candidate, 0, len(stores))
	for _, store := range stores {
		if store.Longitude == 0 && store.Latitude == 0 {
			continue
		}
		distance := haversine(longitude, latitude, store.Longitude, store.Latitude)
		if distance <= distanceKm {
			candidates = append(candidates, candidate{store: store, distance: distance})
		}
	}
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].distance < candidates[i].distance {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}
	if int32(len(candidates)) > count {
		candidates = candidates[:count]
	}
	out := make([]model.Store, 0, len(candidates))
	distances := make([]float64, 0, len(candidates))
	for _, item := range candidates {
		out = append(out, item.store)
		distances = append(distances, item.distance)
	}
	return out, distances, nil
}

// SearchWithDistance is the customer-facing search contract: only available
// stores are considered and results are ordered by distance from the caller.
func (s *StoreService) SearchWithDistance(ctx context.Context, longitude, latitude float64, filter repository.StoreBusinessSearchFilter) ([]model.Store, []float64, error) {
	if !validCoordinate(longitude, latitude) {
		return nil, nil, ErrInvalid
	}
	stores, err := s.repo.Search(ctx, filter)
	if err != nil {
		return nil, nil, err
	}
	type candidate struct {
		store    model.Store
		distance float64
	}
	candidates := make([]candidate, 0, len(stores))
	for _, store := range stores {
		if store.Longitude == 0 && store.Latitude == 0 {
			continue
		}
		candidates = append(candidates, candidate{
			store:    store,
			distance: haversine(longitude, latitude, store.Longitude, store.Latitude),
		})
	}
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].distance < candidates[i].distance {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}
	result := make([]model.Store, 0, len(candidates))
	distances := make([]float64, 0, len(candidates))
	for _, item := range candidates {
		result = append(result, item.store)
		distances = append(distances, item.distance)
	}
	return result, distances, nil
}

func (s *StoreService) GetWithDistance(ctx context.Context, id uint64, longitude, latitude float64) (*model.Store, float64, error) {
	if !validCoordinate(longitude, latitude) {
		return nil, 0, ErrInvalid
	}
	store, err := s.Get(ctx, id)
	if err != nil {
		return nil, 0, err
	}
	if !store.Available() {
		return nil, 0, ErrNotFound
	}
	if store.Longitude == 0 && store.Latitude == 0 {
		return store, 0, nil
	}
	return store, haversine(longitude, latitude, store.Longitude, store.Latitude), nil
}

func (s *StoreService) UpdateProfile(ctx context.Context, id uint64, fields map[string]any) (*model.Store, error) {
	if id == 0 {
		return nil, ErrInvalid
	}
	allowed := map[string]bool{"name": true, "phone": true, "business_hours": true, "email": true}
	for key := range fields {
		if !allowed[key] {
			return nil, ErrInvalid
		}
	}
	return s.update(ctx, id, fields)
}

func (s *StoreService) UpdateLocation(ctx context.Context, id uint64, fields map[string]any) (*model.Store, error) {
	if id == 0 {
		return nil, ErrInvalid
	}
	allowed := map[string]bool{"province": true, "city": true, "district": true, "address": true, "longitude": true, "latitude": true}
	for key := range fields {
		if !allowed[key] {
			return nil, ErrInvalid
		}
	}
	longitude, hasLongitude := numericField(fields["longitude"])
	latitude, hasLatitude := numericField(fields["latitude"])
	if hasLongitude != hasLatitude {
		return nil, ErrInvalid
	}
	if hasLongitude && !validCoordinate(longitude, latitude) {
		return nil, ErrInvalid
	}
	return s.update(ctx, id, fields)
}

func (s *StoreService) UpdateStatus(ctx context.Context, id uint64, status int32) (*model.Store, error) {
	if !validStatus(status) {
		return nil, ErrInvalid
	}
	return s.update(ctx, id, map[string]any{"status": status})
}

func (s *StoreService) ResetPassword(ctx context.Context, id uint64, oldPassword, newPassword string) (*model.Store, string, error) {
	store, err := s.Get(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if oldPassword != "" && !verifyPassword(oldPassword, store.PasswordHash) {
		return nil, "", ErrPassword
	}
	if newPassword == "" {
		newPassword = randomPassword()
	}
	updated, err := s.update(ctx, id, map[string]any{"password": hashPassword(newPassword)})
	return updated, newPassword, err
}

func (s *StoreService) ChangePassword(ctx context.Context, id uint64, oldPassword, newPassword string) (*model.Store, error) {
	if oldPassword == "" || newPassword == "" || oldPassword == newPassword {
		return nil, ErrInvalid
	}
	store, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !verifyPassword(oldPassword, store.PasswordHash) {
		return nil, ErrPassword
	}
	return s.update(ctx, id, map[string]any{"password": hashPassword(newPassword)})
}

func (s *StoreService) update(ctx context.Context, id uint64, fields map[string]any) (*model.Store, error) {
	updated, err := s.repo.Update(ctx, id, fields)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrNotFound
	}
	return updated, nil
}

func validStatus(value int32) bool {
	return value >= StatusClosed && value <= StatusDefunct
}

func validCoordinate(longitude, latitude float64) bool {
	return !math.IsNaN(longitude) && !math.IsInf(longitude, 0) && longitude >= -180 && longitude <= 180 &&
		!math.IsNaN(latitude) && !math.IsInf(latitude, 0) && latitude >= -90 && latitude <= 90
}

func validCoordinates(longitude, latitude float64) bool {
	if longitude == 0 && latitude == 0 {
		return true
	}
	return validCoordinate(longitude, latitude)
}

func numericField(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	default:
		return 0, false
	}
}

func randomPassword() string {
	buf := make([]byte, 9)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("生成初始密码失败: %w", err))
	}
	return hex.EncodeToString(buf)
}

func hashPassword(password string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		panic(fmt.Errorf("生成密码盐失败: %w", err))
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	encode := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=4$%s$%s", encode.EncodeToString(salt), encode.EncodeToString(hash))
}

func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}
	params := map[string]uint32{}
	for _, item := range strings.Split(parts[3], ",") {
		kv := strings.SplitN(item, "=", 2)
		if len(kv) != 2 {
			return false
		}
		value, err := strconv.ParseUint(kv[1], 10, 32)
		if err != nil {
			return false
		}
		params[kv[0]] = uint32(value)
	}
	salt, err1 := decodeHashPart(parts[4])
	expected, err2 := decodeHashPart(parts[5])
	if err1 != nil || err2 != nil {
		return false
	}
	memory, iterations, parallelism := params["m"], params["t"], params["p"]
	if memory == 0 || iterations == 0 || parallelism == 0 || len(expected) == 0 {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, iterations, memory, uint8(parallelism), uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func decodeHashPart(value string) ([]byte, error) {
	if decoded, err := base64.RawStdEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	// 兼容早期骨架版本生成的十六进制 Argon2 格式。
	return hex.DecodeString(value)
}

func haversine(lon1, lat1, lon2, lat2 float64) float64 {
	const earthRadiusKm = 6371.0088
	lat1, lat2 = lat1*math.Pi/180, lat2*math.Pi/180
	dLat := lat2 - lat1
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * earthRadiusKm * math.Asin(math.Sqrt(a))
}

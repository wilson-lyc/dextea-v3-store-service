package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/wilson-lyc/dextea-store-service/internal/model"
)

type StoreRepository interface {
	Create(context.Context, *model.Store) (uint64, error)
	FindByID(context.Context, uint64) (*model.Store, error)
	FindByAccount(context.Context, string) (*model.Store, error)
	FindByIDs(context.Context, []uint64, bool) ([]model.Store, error)
	List(context.Context, int32, int32, StoreListFilter) ([]model.Store, int64, error)
	Search(context.Context, StoreBusinessSearchFilter) ([]model.Store, error)
	SearchAdmin(context.Context, StoreAdminSearchFilter) ([]model.Store, error)
	ListCities(context.Context) ([]string, error)
	CountByStatus(context.Context) ([]model.StoreStatusCount, error)
	Update(context.Context, uint64, map[string]any) (*model.Store, error)
}

type StoreListFilter struct {
	Keyword  string
	Province string
	City     string
	District string
	Status   *int32
}

type StoreBusinessSearchFilter struct {
	Province string
	City     string
	District string
	Keyword  string
}

type StoreAdminSearchFilter struct {
	Province           string
	City               string
	District           string
	Keyword            string
	IncludeUnavailable bool
	Status             *int32
}

type MySQLStoreRepository struct{ db *sqlx.DB }

func NewStoreRepository(db *sqlx.DB) StoreRepository { return &MySQLStoreRepository{db: db} }

const storeColumns = "id, name, province, city, district, address, status, business_hours, phone, longitude, latitude, account, password, email, created_at, updated_at"

func (r *MySQLStoreRepository) Create(ctx context.Context, s *model.Store) (uint64, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO stores (name, province, city, district, address, status, business_hours, phone, longitude, latitude, account, password, email) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		s.Name, s.Province, s.City, s.District, s.Address, s.Status, s.BusinessHours, s.Phone, s.Longitude, s.Latitude, s.Account, s.PasswordHash, s.Email)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return uint64(id), err
}

func (r *MySQLStoreRepository) FindByID(ctx context.Context, id uint64) (*model.Store, error) {
	return r.findOne(ctx, "SELECT "+storeColumns+" FROM stores WHERE id = ?", id)
}

func (r *MySQLStoreRepository) FindByAccount(ctx context.Context, account string) (*model.Store, error) {
	return r.findOne(ctx, "SELECT "+storeColumns+" FROM stores WHERE account = ?", account)
}

func (r *MySQLStoreRepository) FindByIDs(ctx context.Context, ids []uint64, includeUnavailable bool) ([]model.Store, error) {
	if len(ids) == 0 {
		return []model.Store{}, nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids)+2)
	for _, id := range ids {
		args = append(args, id)
	}
	query := "SELECT " + storeColumns + " FROM stores WHERE id IN (" + placeholders + ")"
	if !includeUnavailable {
		query += " AND status IN (0, 1)"
	}
	query += " ORDER BY id"
	var stores []model.Store
	if err := r.db.SelectContext(ctx, &stores, query, args...); err != nil {
		return nil, err
	}
	return stores, nil
}

func (r *MySQLStoreRepository) findOne(ctx context.Context, query string, args ...any) (*model.Store, error) {
	var store model.Store
	if err := r.db.GetContext(ctx, &store, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &store, nil
}

func (r *MySQLStoreRepository) List(ctx context.Context, page, pageSize int32, filter StoreListFilter) ([]model.Store, int64, error) {
	where, args := buildStoreWhere(filter.Keyword, filter.Province, filter.City, filter.District, filter.Status, true)
	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM stores"+where, args...); err != nil {
		return nil, 0, err
	}
	var stores []model.Store
	query := "SELECT " + storeColumns + " FROM stores" + where + " ORDER BY id LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)
	if err := r.db.SelectContext(ctx, &stores, query, args...); err != nil {
		return nil, 0, err
	}
	return stores, total, nil
}

func (r *MySQLStoreRepository) Search(ctx context.Context, filter StoreBusinessSearchFilter) ([]model.Store, error) {
	where, args := buildStoreWhere(filter.Keyword, filter.Province, filter.City, filter.District, nil, false)
	var stores []model.Store
	if err := r.db.SelectContext(ctx, &stores, "SELECT "+storeColumns+" FROM stores"+where+" ORDER BY id", args...); err != nil {
		return nil, err
	}
	return stores, nil
}

func (r *MySQLStoreRepository) SearchAdmin(ctx context.Context, filter StoreAdminSearchFilter) ([]model.Store, error) {
	where, args := buildStoreWhere(filter.Keyword, filter.Province, filter.City, filter.District, filter.Status, filter.IncludeUnavailable)
	var stores []model.Store
	if err := r.db.SelectContext(ctx, &stores, "SELECT "+storeColumns+" FROM stores"+where+" ORDER BY id", args...); err != nil {
		return nil, err
	}
	return stores, nil
}

func (r *MySQLStoreRepository) ListCities(ctx context.Context) ([]string, error) {
	where := " WHERE city <> '' AND status IN (0, 1)"
	var cities []string
	if err := r.db.SelectContext(ctx, &cities, "SELECT DISTINCT city FROM stores"+where+" ORDER BY city"); err != nil {
		return nil, err
	}
	return cities, nil
}

func (r *MySQLStoreRepository) CountByStatus(ctx context.Context) ([]model.StoreStatusCount, error) {
	var counts []model.StoreStatusCount
	if err := r.db.SelectContext(ctx, &counts, "SELECT status, COUNT(*) AS count FROM stores GROUP BY status ORDER BY status"); err != nil {
		return nil, err
	}
	return counts, nil
}

func (r *MySQLStoreRepository) Update(ctx context.Context, id uint64, fields map[string]any) (*model.Store, error) {
	if len(fields) == 0 {
		return r.FindByID(ctx, id)
	}
	keys := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+1)
	for _, key := range sortedKeys(fields) {
		keys = append(keys, key+" = ?")
		args = append(args, fields[key])
	}
	args = append(args, id)
	if _, err := r.db.ExecContext(ctx, "UPDATE stores SET "+strings.Join(keys, ", ")+" WHERE id = ?", args...); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func buildStoreWhere(keyword, province, city, district string, status *int32, includeUnavailable bool) (string, []any) {
	conditions := make([]string, 0, 6)
	args := make([]any, 0, 8)
	if value := strings.TrimSpace(keyword); value != "" {
		pattern := "%" + escapeLike(value) + "%"
		conditions = append(conditions, "(name LIKE ? ESCAPE '\\' OR phone LIKE ? ESCAPE '\\' OR address LIKE ? ESCAPE '\\')")
		args = append(args, pattern, pattern, pattern)
	}
	for _, item := range []struct {
		column string
		value  string
	}{
		{column: "province", value: province},
		{column: "city", value: city},
		{column: "district", value: district},
	} {
		if value := strings.TrimSpace(item.value); value != "" {
			conditions = append(conditions, fmt.Sprintf("%s = ?", item.column))
			args = append(args, value)
		}
	}
	if status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *status)
	} else if !includeUnavailable {
		conditions = append(conditions, "status IN (0, 1)")
	}
	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func escapeLike(value string) string {
	return strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(value)
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

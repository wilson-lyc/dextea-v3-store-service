package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/wilson-lyc/dextea-store-service/internal/model"
)

type StoreRepository interface {
	Create(context.Context, *model.Store) (uint64, error)
	FindByID(context.Context, uint64) (*model.Store, error)
	FindByAccount(context.Context, string) (*model.Store, error)
	List(context.Context, int32, int32, string, *int32) ([]model.Store, int64, error)
	Search(context.Context, string, string, bool) ([]model.Store, error)
	Update(context.Context, uint64, map[string]any) (*model.Store, error)
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

func (r *MySQLStoreRepository) List(ctx context.Context, page, pageSize int32, keyword string, status *int32) ([]model.Store, int64, error) {
	where, args := filter(keyword, status)
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

func (r *MySQLStoreRepository) Search(ctx context.Context, city, keyword string, includeUnavailable bool) ([]model.Store, error) {
	conditions := make([]string, 0, 3)
	args := make([]any, 0, 4)
	if strings.TrimSpace(city) != "" {
		conditions = append(conditions, "city = ?")
		args = append(args, strings.TrimSpace(city))
	}
	if strings.TrimSpace(keyword) != "" {
		pattern := "%" + escapeLike(strings.TrimSpace(keyword)) + "%"
		conditions = append(conditions, "(name LIKE ? ESCAPE '\\' OR address LIKE ? ESCAPE '\\')")
		args = append(args, pattern, pattern)
	}
	if !includeUnavailable {
		conditions = append(conditions, "status IN (0, 1)")
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	var stores []model.Store
	if err := r.db.SelectContext(ctx, &stores, "SELECT "+storeColumns+" FROM stores"+where+" ORDER BY id", args...); err != nil {
		return nil, err
	}
	return stores, nil
}

func (r *MySQLStoreRepository) Update(ctx context.Context, id uint64, fields map[string]any) (*model.Store, error) {
	if len(fields) == 0 {
		return r.FindByID(ctx, id)
	}
	keys := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+1)
	for key, value := range fields {
		keys = append(keys, key+" = ?")
		args = append(args, value)
	}
	args = append(args, id)
	if _, err := r.db.ExecContext(ctx, "UPDATE stores SET "+strings.Join(keys, ", ")+" WHERE id = ?", args...); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func filter(keyword string, status *int32) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 4)
	if strings.TrimSpace(keyword) != "" {
		pattern := "%" + escapeLike(strings.TrimSpace(keyword)) + "%"
		conditions = append(conditions, "(name LIKE ? ESCAPE '\\' OR phone LIKE ? ESCAPE '\\' OR address LIKE ? ESCAPE '\\')")
		args = append(args, pattern, pattern, pattern)
	}
	if status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *status)
	}
	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func escapeLike(value string) string {
	return strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(value)
}

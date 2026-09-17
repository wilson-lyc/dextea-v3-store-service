package repository

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/wilson-lyc/dextea-store-service/internal/config"
)

func NewMySQL(cfg config.MySQLConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("连接数据库 %s:%d/%s 失败: %w", cfg.Host, cfg.Port, cfg.Database, err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return db, nil
}

package model

import "time"

type Store struct {
	ID            uint64    `db:"id"`
	Name          string    `db:"name"`
	Province      string    `db:"province"`
	City          string    `db:"city"`
	District      string    `db:"district"`
	Address       string    `db:"address"`
	Status        int32     `db:"status"`
	BusinessHours string    `db:"business_hours"`
	Phone         string    `db:"phone"`
	Longitude     float64   `db:"longitude"`
	Latitude      float64   `db:"latitude"`
	Account       string    `db:"account"`
	PasswordHash  string    `db:"password"`
	Email         string    `db:"email"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (s Store) Available() bool { return s.Status == 0 || s.Status == 1 }

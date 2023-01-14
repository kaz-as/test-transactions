package domain

import (
	"context"
	"database/sql"
)

type User struct {
	ID      UserID
	Balance Balance
}

type (
	UserID  string
	Balance int64
)

type UsersRepository interface {
	Store(ctx context.Context, tx *sql.Tx, user *User) error
	GetForUpdate(ctx context.Context, tx *sql.Tx, userID UserID) (*User, error)
	Update(ctx context.Context, tx *sql.Tx, user *User) error
}

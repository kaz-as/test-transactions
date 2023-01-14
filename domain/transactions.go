package domain

import (
	"context"
	"database/sql"
	"time"
)

type Tx struct {
	ID        string
	From      UserID
	To        UserID
	Value     Balance
	Timestamp time.Time
}

type TxRepository interface {
	Store(ctx context.Context, tx *sql.Tx, transaction *Tx) error
}

package postgres

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/kaz-as/test-transactions/domain"
	"github.com/kaz-as/test-transactions/pkg/logger"
)

type txRepo struct {
	log logger.Interface
}

func NewRepo(log logger.Interface) domain.TxRepository {
	return &txRepo{
		log: log,
	}
}

func (t *txRepo) Store(ctx context.Context, tx *sql.Tx, transaction *domain.Tx) error {
	query := `INSERT INTO transactions (id, "from", "to", value, timestamp) VALUES (?, ?, ?, ?, ?)`

	src := rand.New(rand.NewSource(time.Now().UnixNano() + int64(transaction.Value)))
	uid, err := generateUID(src)
	if err != nil {
		return fmt.Errorf("generate uid: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer func() {
		err := stmt.Close()
		if err != nil {
			t.log.Error("close stmt: %s", err)
		}
	}()

	_, err = stmt.ExecContext(ctx, uid, string(transaction.From), string(transaction.To), int64(transaction.Value), time.Now())
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

func generateUID(src *rand.Rand) (string, error) {
	b := make([]byte, 32)

	if _, err := src.Read(b); err != nil {
		return "", err
	}

	encoded := hex.EncodeToString(b)
	return encoded[:64], nil
}

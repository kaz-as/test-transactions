package usersrepo

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

type userRepo struct {
	log logger.Interface
}

func NewRepo(log logger.Interface) domain.UsersRepository {
	return &userRepo{
		log: log,
	}
}

func (u *userRepo) Store(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	query := `INSERT INTO users (id, balance) VALUES ($1, $2)`

	src := rand.New(rand.NewSource(time.Now().UnixNano() + int64(user.Balance)))
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
			u.log.Error("close stmt: %s", err)
		}
	}()

	_, err = stmt.ExecContext(ctx, string(uid), int64(user.Balance))
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	user.ID = uid
	return nil
}

func (u *userRepo) GetForUpdate(ctx context.Context, tx *sql.Tx, userID domain.UserID) (*domain.User, error) {
	query := `SELECT * FROM users WHERE id = $1 FOR NO KEY UPDATE`

	row := tx.QueryRowContext(ctx, query, string(userID))

	user := domain.User{}
	err := row.Scan((*string)(&user.ID), (*int64)(&user.Balance))
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return &user, nil
}

func (u *userRepo) Update(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	query := `UPDATE users SET balance = $1 WHERE id = $2`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer func() {
		err := stmt.Close()
		if err != nil {
			u.log.Error("close stmt: %s", err)
		}
	}()

	res, err := stmt.ExecContext(ctx, int64(user.Balance), string(user.ID))
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if affected, _ := res.RowsAffected(); affected != 1 {
		return fmt.Errorf("wierd behaviour: affected %d rows", affected)
	}

	return nil
}

func generateUID(src *rand.Rand) (domain.UserID, error) {
	b := make([]byte, 16)

	if _, err := src.Read(b); err != nil {
		return "", err
	}

	encoded := hex.EncodeToString(b)
	return domain.UserID(encoded[:32]), nil
}

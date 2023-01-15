package general

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/kaz-as/test-transactions/domain"
	"github.com/kaz-as/test-transactions/pkg/logger"
	"time"
)

const PrimaryUserID = "00000000000000000000000000000000"

type UseCase struct {
	logger     logger.Interface
	db         *sql.DB
	usersRepo  domain.UsersRepository
	txRepo     domain.TxRepository
	ctxTimeout time.Duration
}

func NewUseCase(
	log logger.Interface,
	db *sql.DB,
	usersRepo domain.UsersRepository,
	txRepo domain.TxRepository,
	ctxTimeout time.Duration,
) *UseCase {
	return &UseCase{
		logger:     log,
		db:         db,
		usersRepo:  usersRepo,
		txRepo:     txRepo,
		ctxTimeout: ctxTimeout,
	}
}

func (u *UseCase) CreateUser(ctx context.Context, user *domain.User) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, u.ctxTimeout)
	defer cancel()

	tx, err := u.db.BeginTx(ctxTimeout, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer u.rollback(tx)

	err = u.usersRepo.Store(ctxTimeout, tx, user)
	if err != nil {
		return fmt.Errorf("storing user: %w", err)
	}

	businessTx := &domain.Tx{
		From:  PrimaryUserID,
		To:    user.ID,
		Value: user.Balance,
	}

	err = u.txRepo.Store(ctxTimeout, tx, businessTx)
	if err != nil {
		return fmt.Errorf("storing first tx for the user: %w", err)
	}

	return tx.Commit()
}

func (u *UseCase) CreateTx(ctx context.Context, tx *domain.Tx) (
	newBalanceFrom domain.Balance,
	newBalanceTo domain.Balance,
	err error,
) {
	ctxTimeout, cancel := context.WithTimeout(ctx, u.ctxTimeout)
	defer cancel()

	dbTx, err := u.db.BeginTx(ctxTimeout, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer u.rollback(dbTx)

	from, err := u.usersRepo.GetForUpdate(ctxTimeout, dbTx, tx.From)
	if err != nil {
		return 0, 0, fmt.Errorf("get user (from) for update: %w", err)
	}
	to, err := u.usersRepo.GetForUpdate(ctxTimeout, dbTx, tx.To)
	if err != nil {
		return 0, 0, fmt.Errorf("get user (to) for update: %w", err)
	}

	err = u.txRepo.Store(ctxTimeout, dbTx, tx)
	if err != nil {
		return 0, 0, fmt.Errorf("transaction storing: %w", err)
	}

	if err = u.checkBusinessTx(tx, from, to); err != nil {
		return 0, 0, fmt.Errorf("check failed: %w", err)
	}

	from.Balance -= tx.Value
	to.Balance += tx.Value

	err = u.usersRepo.Update(ctxTimeout, dbTx, from)
	if err != nil {
		return 0, 0, fmt.Errorf("update user (from): %w", err)
	}

	err = u.usersRepo.Update(ctxTimeout, dbTx, to)
	if err != nil {
		return 0, 0, fmt.Errorf("update user (to): %w", err)
	}

	return from.Balance, to.Balance, dbTx.Commit()
}

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

func (u *UseCase) checkBusinessTx(tx *domain.Tx, from *domain.User, _ *domain.User) error {
	if from.Balance < tx.Value {
		return fmt.Errorf("user id=%s: %w", from.ID, ErrInsufficientBalance)
	}

	return nil
}

func (u *UseCase) rollback(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		u.logger.Error("tx rollback: %s", err)
	}
}

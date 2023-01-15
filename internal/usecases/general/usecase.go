package general

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/kaz-as/test-transactions/domain"
	"github.com/kaz-as/test-transactions/pkg/logger"
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

	dbTx, err := u.db.BeginTx(ctxTimeout, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer u.rollback(dbTx)

	primaryUser, err := u.usersRepo.GetForUpdate(ctxTimeout, dbTx, PrimaryUserID)
	if err != nil {
		return fmt.Errorf("get primary user: %w", err)
	}

	err = u.usersRepo.Store(ctxTimeout, dbTx, user)
	if err != nil {
		return fmt.Errorf("storing user: %w", err)
	}

	// before check, need to set balance as it was before the transaction
	newUserBalance := user.Balance
	user.Balance = 0
	defer func() {
		user.Balance = newUserBalance
	}()

	businessTx := &domain.Tx{
		From:  PrimaryUserID,
		To:    user.ID,
		Value: newUserBalance,
	}

	if err = u.checkBusinessTx(businessTx, primaryUser, user); err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	primaryUser.Balance -= businessTx.Value

	err = u.txRepo.Store(ctxTimeout, dbTx, businessTx)
	if err != nil {
		return fmt.Errorf("storing first tx for the user: %w", err)
	}

	err = u.usersRepo.Update(ctxTimeout, dbTx, primaryUser)
	if err != nil {
		return fmt.Errorf("update primary user: %w", err)
	}

	return dbTx.Commit()
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

	var from, to *domain.User

	// avoid deadlock
	if strings.Compare(string(tx.From), string(tx.To)) < 1 {
		from, err = u.usersRepo.GetForUpdate(ctxTimeout, dbTx, tx.From)
		if err != nil {
			return 0, 0, fmt.Errorf("get user (from) for update: %w", err)
		}
		to, err = u.usersRepo.GetForUpdate(ctxTimeout, dbTx, tx.To)
		if err != nil {
			return 0, 0, fmt.Errorf("get user (to) for update: %w", err)
		}
	} else {
		to, err = u.usersRepo.GetForUpdate(ctxTimeout, dbTx, tx.To)
		if err != nil {
			return 0, 0, fmt.Errorf("get user (to) for update: %w", err)
		}
		from, err = u.usersRepo.GetForUpdate(ctxTimeout, dbTx, tx.From)
		if err != nil {
			return 0, 0, fmt.Errorf("get user (from) for update: %w", err)
		}
	}

	if err = u.checkBusinessTx(tx, from, to); err != nil {
		return 0, 0, fmt.Errorf("check failed: %w", err)
	}

	err = u.txRepo.Store(ctxTimeout, dbTx, tx)
	if err != nil {
		return 0, 0, fmt.Errorf("transaction storing: %w", err)
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
	ErrSame                = errors.New("to = from")
	ErrNegativeTx          = errors.New("negative tx")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrTooMuch             = errors.New("too much")
)

func (u *UseCase) checkBusinessTx(tx *domain.Tx, from *domain.User, to *domain.User) error {
	if from.ID == to.ID {
		return fmt.Errorf("user id=%s: %w", from.ID, ErrSame)
	}
	if tx.Value < 0 {
		return fmt.Errorf("value=%d: %w", tx.Value, ErrNegativeTx)
	}
	if from.Balance < tx.Value {
		return fmt.Errorf("user id=%s: %w", from.ID, ErrInsufficientBalance)
	}
	if math.MaxInt64-tx.Value < to.Balance {
		return fmt.Errorf("user id=%s: %w", to.ID, ErrTooMuch)
	}

	return nil
}

func (u *UseCase) rollback(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		u.logger.Error("tx rollback: %s", err)
	}
}

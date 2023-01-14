package domain

import "context"

type UseCase interface {
	CreateUser(ctx context.Context, user *User) error
	CreateTx(ctx context.Context, tx *Tx) (newBalanceFrom Balance, newBalanceTo Balance, err error)
}

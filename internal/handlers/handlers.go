package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"

	"github.com/kaz-as/test-transactions/domain"
	"github.com/kaz-as/test-transactions/models"
	"github.com/kaz-as/test-transactions/pkg/logger"
	"github.com/kaz-as/test-transactions/restapi"
	"github.com/kaz-as/test-transactions/restapi/operations"
)

type handlerSet struct {
	log logger.Interface
	db  *sql.DB
	uc  domain.UseCase
}

func newHandlerSet(
	log logger.Interface,
	db *sql.DB,
	uc domain.UseCase,
) *handlerSet {
	return &handlerSet{
		log: log,
		db:  db,
		uc:  uc,
	}
}

func (s *handlerSet) CreateUserHandler(params operations.CreateUserParams) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	user := domain.User{
		Balance: domain.Balance(*params.User.Balance),
	}

	err := s.uc.CreateUser(ctx, &user)
	if err != nil {
		s.log.Error("create user failed: %s", err)
		return operations.NewCreateUserDefault(0)
	}

	userID := string(user.ID)

	ret := operations.NewCreateUserOK().WithPayload(&models.CreateUserSuccess{ID: &userID})
	s.log.Info("create user success: %s: %d", userID, *params.User.Balance)
	return ret
}

func (s *handlerSet) CreateTxHandler(params operations.CreateTxParams) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	tx := domain.Tx{
		From:  domain.UserID(*params.Tx.From),
		To:    domain.UserID(*params.Tx.To),
		Value: domain.Balance(*params.Tx.Value),
	}

	newBalanceFrom, newBalanceTo, err := s.uc.CreateTx(ctx, &tx)
	if err != nil {
		s.log.Error("create tx failed: %s", err)
		return operations.NewCreateTxDefault(0)
	}

	ret := operations.NewCreateTxOK().WithPayload(&models.CreateTxSuccess{
		NewBalanceFrom: (*int64)(&newBalanceFrom),
		NewBalanceTo:   (*int64)(&newBalanceTo),
	})

	s.log.Info("create tx success: %s->%s: %d", *params.Tx.From, *params.Tx.To, *params.Tx.Value)
	return ret
}

func New(
	log logger.Interface,
	db *sql.DB,
	uc domain.UseCase,
) (http.Handler, error) {
	swaggerDoc, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("swagger loading: %s", err)
	}

	api := operations.NewTxAPI(swaggerDoc)
	api.Logger = log.Info
	api.UseSwaggerUI()

	hSet := newHandlerSet(log, db, uc)

	api.CreateUserHandler = operations.CreateUserHandlerFunc(hSet.CreateUserHandler)
	api.CreateTxHandler = operations.CreateTxHandlerFunc(hSet.CreateTxHandler)

	return api.Serve(middleware.Builder(nil)), nil
}

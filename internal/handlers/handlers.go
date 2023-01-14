package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"

	"github.com/kaz-as/test-transactions/domain"
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

	// todo add exact handlers
	_ = hSet

	return api.Serve(middleware.Builder(nil)), nil
}

package app

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/kaz-as/test-transactions/config"
	"github.com/kaz-as/test-transactions/internal/handlers"
	"github.com/kaz-as/test-transactions/internal/middlewares"
	transactions "github.com/kaz-as/test-transactions/internal/transactions/repository/postgres"
	"github.com/kaz-as/test-transactions/internal/usecases/general"
	users "github.com/kaz-as/test-transactions/internal/users/repository/postgres"
	"github.com/kaz-as/test-transactions/pkg/httpserver"
	"github.com/kaz-as/test-transactions/pkg/logger"
)

type App struct {
	log  logger.Interface
	srv  *httpserver.Server
	conn *sql.DB
}

func New(cfg *config.Config) (app App, _ error) {
	var l logger.Interface = logger.New(cfg.Level)

	app.log = l

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", cfg.DB.User, cfg.DB.Pass, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return app, fmt.Errorf("db open: %s", err)
	}

	app.conn = db

	db.SetConnMaxIdleTime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	err = db.Ping()
	if err != nil {
		return app, fmt.Errorf("db ping: %s", err)
	}

	usersRepo := users.NewRepo(l)
	transactionsRepo := transactions.NewRepo(l)

	usecase := general.NewUseCase(l, db, usersRepo, transactionsRepo, cfg.DB.Timeout)

	h, err := handlers.New(l, db, usecase)
	if err != nil {
		return app, fmt.Errorf("creating main handler: %s", err)
	}

	mwGlobal := middlewares.Chain([]middlewares.Middleware{
		middlewares.Logger(l),
		middlewares.Recoverer(l),
	})

	app.srv = httpserver.New(
		mwGlobal(h),
		httpserver.Port(cfg.Port),
		httpserver.Logger(l),
	)

	return app, nil
}

// Close must be called at app stop or exit
func (app App) Close() {
	if app.conn != nil {
		err := app.conn.Close()
		if err != nil {
			app.log.Error("db connection cannot be closed: %s", err)
		}
	}

	if app.srv != nil {
		err := app.srv.Shutdown()
		if err != nil {
			app.log.Error("srv.Shutdown: %s", err)
		}
	}
}

// Run returns only application error that cause shutdown, else nil
func (app App) Run() (ret error) {
	app.srv.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		app.log.Info("app - Run - signal: " + s.String())
	case err := <-app.srv.Notify():
		ret = fmt.Errorf("srv.Notify: %w", err)
		app.log.Error(ret)
	}

	return
}

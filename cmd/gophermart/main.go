package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/darrior/gophermart/cmd/gophermart/config"
	"github.com/darrior/gophermart/internal/gateways/accrual"
	"github.com/darrior/gophermart/internal/handlers"
	"github.com/darrior/gophermart/internal/repository"
	"github.com/darrior/gophermart/internal/service"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)

	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot parse config")
	}

	db, err := sqlx.Open("pgx", cfg.DatabaseConnConf.ConnString())
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot open DB")
	}
	go func() {
		<-ctx.Done()
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("cannot close DB connection")
		}
	}()

	r, err := repository.NewRepository(ctx, db)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize repository")
	}

	a := accrual.NewAccrual(cfg.AccrualSystemAddress)
	s := service.NewService(ctx, r, a, 5)

	done := make(chan struct{})

	go func() {
		s.StartWorkers(ctx)
		close(done)
	}()

	srv := handlers.NewServer(cfg.RunAddress, s)

	if err := srv.Start(ctx); err != nil {
		log.Error().Err(err).Msg("Unexpected shutdown of server")
	}

	<-done
}

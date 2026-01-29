// Package migrations provides Up and Down functions to migrate DB.
package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/darrior/gophermart/migrations"
	"github.com/pressly/goose/v3"
)

func Up(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("cannot set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("cannot migrate DB: %w", err)
	}

	return nil
}

func Down(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("cannot set dialect: %w", err)
	}

	if err := goose.DownContext(ctx, db, "."); err != nil {
		return fmt.Errorf("cannot migrate DB: %w", err)
	}

	return nil
}

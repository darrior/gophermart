// Package migrations provides embed FS with migration files.
package migrations

import "embed"

//go:embed *.sql
var EmbedMigrations embed.FS

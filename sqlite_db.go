package buna

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type SQLiteDB struct {
	db     *sql.DB
	logger *zap.Logger
}

func OpenSQLiteDB(ctx context.Context, logger *zap.Logger, dsn string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("buna: sqlite_db: failed to open sqlite db: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("buna: sqlite_db: sqlite db down: %w", err)
	}

	s := &SQLiteDB{
		db:     db,
		logger: logger,
	}

	if err := s.migrate(ctx); err != nil {
		s.Close()
		return nil, fmt.Errorf("buna: sqlite_db: failed to migrate SQLite database: %w", err)
	}

	return s, nil
}

func (s *SQLiteDB) migrate(ctx context.Context) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS coffees (
				id INTEGER NOT NULL PRIMARY KEY,
				name TEXT NOT NULL,
				roaster TEXT NULL,
				region TEXT NULL,
				variety TEXT NULL,
				method TEXT NULL,
				decaf BOOLEAN NULL CHECK (decaf IN (0,1)),
				UNIQUE(name, roaster)
			)
		`); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to create coffees table: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: sqlite_db: transaction failed: %w", err)
	}

	return nil
}

func (s *SQLiteDB) TransactContext(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("buna: sqlite_db: failed to begin a transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				s.logger.Error("buna: sqlite_db: transaction rollback failed")
			}
			return
		}

		err = tx.Commit()
	}()

	return f(ctx, tx)
}

func (s *SQLiteDB) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("buna: sqlite_db: failed to close sqlite db: %w", err)
	}
	return nil
}

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
				decaf BOOLEAN NULL
					CHECK (decaf IN (0,1)),
				UNIQUE(name, roaster)
			)
		`); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to create coffees table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS purchases (
				id INTEGER NOT NULL PRIMARY KEY,
				coffee_id INTEGER NOT NULL,
				bought_date TEXT NOT NULL,
				roast_date TEXT NULL,
				FOREIGN KEY (coffee_id)
					REFERENCES coffees (id)
						ON DELETE RESTRICT
			)
		`); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to create purchases table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS brewing_methods (
				id INTEGER NOT NULL PRIMARY KEY,
				name TEXT NOT NULL,
				UNIQUE(name)
			)
		`); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to create brewing_methods table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS grinders (
				id INTEGER NOT NULL PRIMARY KEY,
				name TEXT NOT NULL,
				company TEXT NULL,
				max_grind_setting INTEGER NULL,
				UNIQUE(name)
			)
		`); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to create grinders table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS brewings (
				id INTEGER NOT NULL PRIMARY KEY,
				coffee_id INTEGER NOT NULL,
				method_id INTEGER NOT NULL,
				date TEXT NOT NULL,
				roast_date TEXT NULL,
				grinder_id INTEGER NOT NULL,
				grind_setting INTEGER NOT NULL
					CHECK (grind_setting >= 0),
				total_brewing_time_sec INTEGER NOT NULL
					CHECK (total_brewing_time_sec > 0),
				water_grams INTEGER NOT NULL
					CHECK (water_grams > 0),
				coffee_grams INTEGER NOT NULL
					CHECK (coffee_grams > 0),
				v60_filter_type TEXT NULL
					CHECK (v60_filter_type IN ("", "eu", "jp")),
				rating INTEGER NULL
					CHECK (rating >= 0 AND rating <= 10),
				recommended_grind_setting_adjustment TEXT NULL
					CHECK (recommended_grind_setting_adjustment IN ("", "lower", "higher")),
				recommended_coffee_weight_adjustment_grams INTEGER NULL,
				notes TEXT NULL,
				FOREIGN KEY (coffee_id)
					REFERENCES coffees (id)
						ON DELETE RESTRICT,
				FOREIGN KEY (method_id)
					REFERENCES brewing_methods (id)
						ON DELETE RESTRICT,
				FOREIGN KEY (grinder_id)
					REFERENCES grinders (id)
						ON DELETE RESTRICT
			)
		`); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to create brewings table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS cuppings (
				id INTEGER NOT NULL PRIMARY KEY,
				date TEXT NOT NULL,
				duration_min INTEGER NOT NULL
					CHECK (duration_min > 0),
				notes TEXT NOT NULL
			)
		`); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to create cuppings table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS cupped_coffees (
				cupping_id INTEGER NOT NULL,
				coffee_id INTEGER NOT NULL,
				rank INTEGER NOT NULL
					CHECK (rank > 0),
				notes TEXT NOT NULL,
				PRIMARY KEY (cupping_id, coffee_id),
				FOREIGN KEY (cupping_id)
					REFERENCES cuppings (id)
						ON DELETE RESTRICT,
				FOREIGN KEY (coffee_id)
					REFERENCES coffees (id)
						ON DELETE RESTRICT
			)
		`); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to create cupped_coffees table: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: sqlite_db: transaction failed: %w", err)
	}

	return nil
}

func (s *SQLiteDB) TransactContext(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) (err error) {
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

package buna

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type dbEntity int

const (
	brewings dbEntity = iota
	brewingMethods
	coffees
	coffeePurchases
	cuppings
	grinders
)

var (
	dbEntityToStringMap = map[dbEntity]string{
		brewings:        "brewings",
		brewingMethods:  "brewing_methods",
		coffees:         "coffees",
		coffeePurchases: "purchases",
		cuppings:        "cuppings",
		grinders:        "grinders",
	}

	dbEntityToName = map[dbEntity]string{
		brewings:        "brewings",
		brewingMethods:  "brewing methods",
		coffees:         "coffees",
		coffeePurchases: "coffee purchases",
		cuppings:        "cuppings",
		grinders:        "grinders",
	}
)

func (s *SQLiteDB) getTotalCount(ctx context.Context, entity dbEntity) (int, error) {
	dbEntityString, ok := dbEntityToStringMap[entity]
	if !ok {
		return 0, fmt.Errorf("buna: sqlite_db_statistics: unable to map dbEntity to string")
	}

	var count int
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		query := `
			SELECT count(*)
			FROM ?
		`
		query = strings.Replace(query, "?", dbEntityString, 1)
		if err := tx.QueryRowContext(ctx, query, sql.Named("entity", dbEntityString)).Scan(&count); err != nil {
			return fmt.Errorf("buna: sqlite_db_statistics: failed to retrieve total count from db: %w", err)
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("buna: sqlite_db_statistics: getTotalCount transaction failed: %w", err)
	}

	return count, nil
}

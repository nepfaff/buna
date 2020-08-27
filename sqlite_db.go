package buna

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

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
				roaster TEXT NOT NULL,
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

func (s *SQLiteDB) getBrewingsByLastAdded(ctx context.Context, limit int) ([]brewing, error) {
	brewings := make([]brewing, 0, limit)
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT 	b.date,
					c.name,
					c.roaster,
					m.name,
					b.roast_date,
					g.name,
					b.grind_setting,
					b.total_brewing_time_sec,
					b.coffee_grams,
					b.water_grams,
					b.v60_filter_type,
					b.rating,
					b.recommended_grind_setting_adjustment,
					b.recommended_coffee_weight_adjustment_grams,
					b.notes
			FROM brewings AS b
			INNER JOIN coffees AS c
				ON c.id = b.coffee_id
			INNER JOIN brewing_methods AS m
				ON m.id = b.method_id
			INNER JOIN grinders AS g
				ON g.id = b.grinder_id
			ORDER BY b.id DESC
			LIMIT :limit
		`,
			sql.Named("limit", limit),
		)
		if err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to retrieve brewing rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var brewing brewing
			var roastDate, v60FilterType, rating, recommendedGrindSettingAdjustment, recommendedCoffeeWeightAdjustmentGrams, notes interface{}
			if err := rows.Scan(
				&brewing.date,
				&brewing.coffeeName,
				&brewing.coffeeRoaster,
				&brewing.brewingMethodName,
				&roastDate,
				&brewing.grinderName,
				&brewing.grindSetting,
				&brewing.totalBrewingTimeSec,
				&brewing.coffeeGrams,
				&brewing.waterGrams,
				&v60FilterType,
				&rating,
				&recommendedGrindSettingAdjustment,
				&recommendedCoffeeWeightAdjustmentGrams,
				&notes,
			); err != nil {
				return fmt.Errorf("buna: sqlite_db: failed to scan row: %w", err)
			}

			// Deal with possible NULL values
			if v := reflect.ValueOf(roastDate); v.Kind() == reflect.String {
				brewing.roastDate = roastDate.(string)
			} else {
				brewing.roastDate = "Unknown"
			}
			if v := reflect.ValueOf(v60FilterType); v.Kind() == reflect.String {
				brewing.v60FilterType = v60FilterType.(string)
			} else {
				brewing.v60FilterType = "None"
			}
			if v := reflect.ValueOf(rating); v.Kind() == reflect.Int64 {
				brewing.rating = int(rating.(int64))
			} else {
				brewing.rating = 0
			}
			if v := reflect.ValueOf(recommendedGrindSettingAdjustment); v.Kind() == reflect.String {
				brewing.recommendedGrindSettingAdjustment = recommendedGrindSettingAdjustment.(string)
			} else {
				brewing.recommendedGrindSettingAdjustment = "None"
			}
			if v := reflect.ValueOf(recommendedCoffeeWeightAdjustmentGrams); v.Kind() == reflect.Int64 {
				brewing.recommendedCoffeeWeightAdjustmentGrams = int(recommendedCoffeeWeightAdjustmentGrams.(int64))
			} else {
				brewing.recommendedCoffeeWeightAdjustmentGrams = 0
			}
			if v := reflect.ValueOf(notes); v.Kind() == reflect.String {
				brewing.notes = notes.(string)
			} else {
				brewing.notes = "None"
			}

			brewings = append(brewings, brewing)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: sqlite_db: displayBrewingsByLastAdded transaction failed: %w", err)
	}

	return brewings, nil
}

// The following fields are used from the brewingFilter argument:
// brewingMethodName, v60FilterType, coffeeName, coffeeRoaster, coffeeGrams, waterGrams, grinderName
func (s *SQLiteDB) getBrewingSuggestions(ctx context.Context, limit int, brewingFilter brewing) ([]brewing, error) {
	brewings := make([]brewing, 0, limit)
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT 	b.grind_setting,
					b.total_brewing_time_sec,
					b.coffee_grams,
					b.water_grams,
					b.recommended_grind_setting_adjustment,
					b.recommended_coffee_weight_adjustment_grams,
					b.rating,
					b.v60_filter_type,
					c.name,
					b.date,
					g.name,
					b.notes
			FROM brewings AS b
			INNER JOIN coffees AS c
				ON c.id = b.coffee_id
			INNER JOIN brewing_methods AS m
				ON m.id = b.method_id
			INNER JOIN grinders AS g
				ON g.id = b.grinder_id
			WHERE (m.name = :brewingMethodName)
			AND (b.v60_filter_type = :v60FilterType OR "" = :v60FilterType)
			AND (c.name = :coffeeName OR "" = :coffeeName)
			AND (c.roaster = :coffeeRoaster OR "" = :coffeeRoaster)
			AND (b.coffee_grams = :coffeeGrams OR 0 = :coffeeGrams)
			AND (b.water_grams = :waterGrams OR 0 = :waterGrams)
			AND (g.name = :grinderName OR "" = :grinderName)
			ORDER BY b.id DESC
			LIMIT :limit
		`,
			sql.Named("limit", limit),
			sql.Named("brewingMethodName", brewingFilter.brewingMethodName),
			sql.Named("v60FilterType", brewingFilter.v60FilterType),
			sql.Named("coffeeName", brewingFilter.coffeeName),
			sql.Named("coffeeRoaster", brewingFilter.coffeeRoaster),
			sql.Named("coffeeGrams", brewingFilter.coffeeGrams),
			sql.Named("waterGrams", brewingFilter.waterGrams),
			sql.Named("grinderName", brewingFilter.grinderName),
		)
		if err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to retrieve brewing suggestion rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var brewing brewing
			var recommendedGrindSettingAdjustment, recommendedCoffeeWeightAdjustmentGrams, rating, v60FilterType, notes interface{}
			if err := rows.Scan(
				&brewing.grindSetting,
				&brewing.totalBrewingTimeSec,
				&brewing.coffeeGrams,
				&brewing.waterGrams,
				&recommendedGrindSettingAdjustment,
				&recommendedCoffeeWeightAdjustmentGrams,
				&rating,
				&v60FilterType,
				&brewing.coffeeName,
				&brewing.date,
				&brewing.grinderName,
				&notes,
			); err != nil {
				return fmt.Errorf("buna: sqlite_db: failed to scan row: %w", err)
			}

			// Deal with possible NULL values
			if v := reflect.ValueOf(recommendedGrindSettingAdjustment); v.Kind() == reflect.String {
				brewing.recommendedGrindSettingAdjustment = recommendedGrindSettingAdjustment.(string)
			} else {
				brewing.recommendedGrindSettingAdjustment = "None"
			}
			if v := reflect.ValueOf(recommendedCoffeeWeightAdjustmentGrams); v.Kind() == reflect.Int64 {
				brewing.recommendedCoffeeWeightAdjustmentGrams = int(recommendedCoffeeWeightAdjustmentGrams.(int64))
			} else {
				brewing.recommendedCoffeeWeightAdjustmentGrams = 0
			}
			if v := reflect.ValueOf(rating); v.Kind() == reflect.Int64 {
				brewing.rating = int(rating.(int64))
			} else {
				brewing.rating = 0
			}
			if v := reflect.ValueOf(v60FilterType); v.Kind() == reflect.String {
				brewing.v60FilterType = v60FilterType.(string)
			} else {
				brewing.v60FilterType = "None"
			}
			if v := reflect.ValueOf(notes); v.Kind() == reflect.String {
				brewing.notes = notes.(string)
			} else {
				brewing.notes = "None"
			}

			brewings = append(brewings, brewing)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: sqlite_db: getBrewingSuggestions transaction failed: %w", err)
	}

	return brewings, nil
}

func (s *SQLiteDB) getCoffeesByLastAdded(ctx context.Context, limit int) ([]coffee, error) {
	coffees := make([]coffee, 0, limit)
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT name, roaster, region, variety, method, decaf
			FROM coffees
			ORDER BY id DESC
			LIMIT :limit
		`,
			sql.Named("limit", limit),
		)
		if err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to retrieve coffee rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var coffee coffee
			var region, variety, method, decaf interface{}
			if err := rows.Scan(&coffee.name, &coffee.roaster, &region, &variety, &method, &decaf); err != nil {
				return fmt.Errorf("buna: sqlite_db: failed to scan row: %w", err)
			}

			// Deal with possible NULL values
			if v := reflect.ValueOf(region); v.Kind() == reflect.String {
				coffee.region = region.(string)
			} else {
				coffee.region = "Unknown"
			}
			if v := reflect.ValueOf(variety); v.Kind() == reflect.String {
				coffee.variety = variety.(string)
			} else {
				coffee.variety = "Unknown"
			}
			if v := reflect.ValueOf(method); v.Kind() == reflect.String {
				coffee.method = method.(string)
			} else {
				coffee.method = "Unknown"
			}
			if v := reflect.ValueOf(decaf); v.Kind() == reflect.Bool {
				coffee.decaf = decaf.(bool)
			}

			coffees = append(coffees, coffee)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: sqlite_db: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: sqlite_db: displayCoffeesByLastAdded transaction failed: %w", err)
	}

	return coffees, nil
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

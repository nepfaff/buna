package buna

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

func (s *SQLiteDB) getBrewingsOrderByDesc(ctx context.Context, limit int, orderByName string) ([]brewing, error) {
	brewings := make([]brewing, 0, limit)
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, fmt.Sprintf(`
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
			ORDER BY b.%s DESC, b.id DESC
			LIMIT :limit
		`, orderByName),
			sql.Named("limit", limit),
		)
		if err != nil {
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to retrieve brewing rows: %w", err)
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
				return fmt.Errorf("buna: sqlite_db_retrieve: failed to scan row: %w", err)
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
			if v := reflect.ValueOf(recommendedCoffeeWeightAdjustmentGrams); v.Kind() == reflect.Float64 {
				brewing.recommendedCoffeeWeightAdjustmentGrams = recommendedCoffeeWeightAdjustmentGrams.(float64)
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
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: sqlite_db_retrieve: getBrewingsByLastAdded transaction failed: %w", err)
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
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to retrieve brewing suggestion rows: %w", err)
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
				return fmt.Errorf("buna: sqlite_db_retrieve: failed to scan row: %w", err)
			}

			// Deal with possible NULL values
			if v := reflect.ValueOf(recommendedGrindSettingAdjustment); v.Kind() == reflect.String {
				brewing.recommendedGrindSettingAdjustment = recommendedGrindSettingAdjustment.(string)
			} else {
				brewing.recommendedGrindSettingAdjustment = "None"
			}
			if v := reflect.ValueOf(recommendedCoffeeWeightAdjustmentGrams); v.Kind() == reflect.Float64 {
				brewing.recommendedCoffeeWeightAdjustmentGrams = recommendedCoffeeWeightAdjustmentGrams.(float64)
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
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: sqlite_db_retrieve: getBrewingSuggestions transaction failed: %w", err)
	}

	return brewings, nil
}

func (s *SQLiteDB) getCoffeeIDByNameRoaster(ctx context.Context, name string, roaster string) (int, error) {
	var coffeeID int
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `
			SELECT id from coffees
			WHERE name = :name AND roaster = :roaster
		`,
			sql.Named("name", name),
			sql.Named("roaster", roaster),
		).Scan(&coffeeID); err != nil {
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to retrieve coffee id from db: %w", err)
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("buna: sqlite_db_retrieve: getCoffeeIDByNameRoaster transaction failed: %w", err)
	}

	return coffeeID, nil
}

func (s *SQLiteDB) getCoffeePurchasesByLastAdded(ctx context.Context, limit int) ([]coffeePurchase, error) {
	coffeePurchases := make([]coffeePurchase, 0, limit)
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT c.name, c.roaster, p.bought_date, p.roast_date
			FROM purchases AS p
			INNER JOIN coffees AS c
				ON p.coffee_id = c.id
			ORDER BY p.id DESC
			LIMIT :limit
		`,
			sql.Named("limit", limit),
		)
		if err != nil {
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to retrieve coffee purchase rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var coffeePurchase coffeePurchase
			var roastDate interface{}
			if err := rows.Scan(&coffeePurchase.coffeeName, &coffeePurchase.coffeeRoaster, &coffeePurchase.boughtDate, &roastDate); err != nil {
				return fmt.Errorf("buna: sqlite_db_retrieve: failed to scan row: %w", err)
			}

			// Deal with possible NULL values
			if v := reflect.ValueOf(roastDate); v.Kind() == reflect.String {
				coffeePurchase.roastDate = roastDate.(string)
			} else {
				coffeePurchase.roastDate = "Unknown"
			}

			coffeePurchases = append(coffeePurchases, coffeePurchase)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: sqlite_db_retrieve: getCoffeePurchasesByLastAdded transaction failed: %w", err)
	}

	return coffeePurchases, nil
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
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to retrieve coffee rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var coffee coffee
			var region, variety, method, decaf interface{}
			if err := rows.Scan(&coffee.name, &coffee.roaster, &region, &variety, &method, &decaf); err != nil {
				return fmt.Errorf("buna: sqlite_db_retrieve: failed to scan row: %w", err)
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
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: sqlite_db_retrieve: getCoffeesByLastAdded transaction failed: %w", err)
	}

	return coffees, nil
}

func (s *SQLiteDB) getCuppingsByLastAdded(ctx context.Context, limit int) ([]cupping, error) {
	var cuppings []cupping
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var realLimit interface{}
		if err := tx.QueryRowContext(ctx, `
			SELECT sum(coffeesPerCupping)
			FROM (
				SELECT cu.id, count(*) as coffeesPerCupping
				FROM cuppings AS cu
				INNER JOIN cupped_coffees AS cc
					ON cu.id = cc.cupping_id
				GROUP BY cu.id
				LIMIT :limit
			)
		`,
			sql.Named("limit", limit),
		).Scan(&realLimit); err != nil {
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to retrieve real cupping limit from db: %w", err)
		}

		var realLimitInt int64
		if v := reflect.ValueOf(realLimit); v.Kind() == reflect.Int64 {
			realLimitInt = realLimit.(int64)
		} else {
			// realLimit == NULL
			return nil
		}

		rows, err := tx.QueryContext(ctx, `
			SELECT cu.id, cu.date, cu.duration_min, cu.notes, c.name, cc.rank, cc.notes
			FROM cuppings AS cu
			INNER JOIN cupped_coffees AS cc
				ON cu.id = cc.cupping_id
			INNER JOIN coffees AS c
				ON c.id = cc.coffee_id
			ORDER BY cu.id DESC, cc.rank
			LIMIT :limit
		`,
			sql.Named("limit", realLimitInt),
		)
		if err != nil {
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to retrieve cupping rows: %w", err)
		}
		defer rows.Close()

		isFirstRow := true
		var prevCuppingID int
		var cupping cupping
		var cuppedCoffees []cuppedCoffee
		for rows.Next() {
			var cuppingID int
			var coffee cuppedCoffee
			if err := rows.Scan(
				&cuppingID,
				&cupping.date,
				&cupping.durationMin,
				&cupping.notes,
				&coffee.name,
				&coffee.rank,
				&coffee.notes,
			); err != nil {
				return fmt.Errorf("buna: sqlite_db_retrieve: failed to scan cupping row: %w", err)
			}

			if isFirstRow {
				prevCuppingID = cuppingID
				isFirstRow = false
			}

			if prevCuppingID != cuppingID {
				cupping.cuppedCoffees = make([]cuppedCoffee, len(cuppedCoffees))
				copy(cupping.cuppedCoffees, cuppedCoffees)
				cuppings = append(cuppings, cupping)

				cuppedCoffees = cuppedCoffees[:0]
				prevCuppingID = cuppingID
			}

			cuppedCoffees = append(cuppedCoffees, coffee)
		}
		cupping.cuppedCoffees = cuppedCoffees
		cuppings = append(cuppings, cupping)

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: sqlite_db_retrieve: getCuppingsByLastAdded transaction failed: %w", err)
	}

	return cuppings, nil
}

func (s *SQLiteDB) getGrinderIDByName(ctx context.Context, name string) (int, error) {
	var grinderID int
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `
			SELECT id
			FROM grinders
			WHERE name = :grinderName
		`,
			sql.Named("grinderName", name),
		).Scan(&grinderID); err != nil {
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to retrieve grinder id from db: %w", err)
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("buna: sqlite_db_retrieve: getGrinderIDByName transaction failed: %w", err)
	}

	return grinderID, nil
}

func (s *SQLiteDB) getMethodIDByName(ctx context.Context, name string) (int, error) {
	var methodID int
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `
			SELECT id
			FROM brewing_methods
			WHERE name = :brewingMethodName
		`,
			sql.Named("brewingMethodName", name),
		).Scan(&methodID); err != nil {
			return fmt.Errorf("buna: sqlite_db_retrieve: failed to retrieve method id from db: %w", err)
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("buna: sqlite_db_retrieve: getMethodIDByName transaction failed: %w", err)
	}

	return methodID, nil
}

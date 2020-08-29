package buna

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *SQLiteDB) insertBrewing(ctx context.Context, brewing brewing) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		coffeeID, err := s.getCoffeeIDByNameRoaster(ctx, brewing.coffeeName, brewing.coffeeRoaster)
		if err != nil {
			fmt.Println("Unable to link this brewing to an existing coffee. Please create a new coffee first and then try again.")
			return nil
		}

		methodID, err := s.getMethodIDByName(ctx, brewing.brewingMethodName)
		if err != nil {
			fmt.Println("Unable to link this brewing to an existing brewing method. Please create a new brewing method first and then try again.")
			return nil
		}

		grinderID, err := s.getGrinderIDByName(ctx, brewing.grinderName)
		if err != nil {
			fmt.Println("Unable to link this brewing to an existing coffee grinder. Please create a new coffee grinder first and then try again.")
			return nil
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO brewings(
				coffee_id,
				method_id,
				grinder_id,
				date,
				roast_date,
				grind_setting,
				total_brewing_time_sec,
				water_grams,
				coffee_grams,
				v60_filter_type,
				rating,
				recommended_grind_setting_adjustment,
				recommended_coffee_weight_adjustment_grams,
				notes
			)
			VALUES (
				:coffeeID,
				:methodID,
				:grinderID,
				:date,
				:roastDate,
				:grindSetting,
				:totalBrewingTimeSec,
				:waterGrams,
				:coffeeGrams,
				:v60FilterType,
				:rating,
				:recommendedGrindSettingAdjustment,
				:recommendedCoffeeWeightAdjustmentGrams,
				:notes
			)
		`,
			sql.Named("coffeeID", coffeeID),
			sql.Named("methodID", methodID),
			sql.Named("grinderID", grinderID),
			sql.Named("date", brewing.date),
			sql.Named("roastDate", brewing.roastDate),
			sql.Named("grindSetting", brewing.grindSetting),
			sql.Named("totalBrewingTimeSec", brewing.totalBrewingTimeSec),
			sql.Named("waterGrams", brewing.waterGrams),
			sql.Named("coffeeGrams", brewing.coffeeGrams),
			sql.Named("v60FilterType", brewing.v60FilterType),
			sql.Named("rating", brewing.rating),
			sql.Named("recommendedGrindSettingAdjustment", brewing.recommendedGrindSettingAdjustment),
			sql.Named("recommendedCoffeeWeightAdjustmentGrams", brewing.recommendedCoffeeWeightAdjustmentGrams),
			sql.Named("notes", brewing.notes),
		); err != nil {
			return fmt.Errorf("buna: brewing: failed to insert coffee brewing into db: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE brewings
			SET roast_date = NULLIF(roast_date, "0-00-00"),
				v60_filter_type = NULLIF(v60_filter_type, ""),
				rating = NULLIF(rating, 0),
				recommended_grind_setting_adjustment = NULLIF(recommended_grind_setting_adjustment, "")
			WHERE coffee_id = :coffeeID AND date = :date
		`,
			sql.Named("coffeeID", coffeeID),
			sql.Named("date", brewing.date),
		); err != nil {
			return fmt.Errorf("buna: brewing: failed to set null values: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: brewing: transaction failed: %w", err)
	}
	return nil
}

func (s *SQLiteDB) insertBrewingMethod(ctx context.Context, brewingMethod brewingMethod) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO brewing_methods(name)
			VALUES (:name)
		`,
			sql.Named("name", brewingMethod.name),
		); err != nil {
			return fmt.Errorf("buna: brewing_method: failed to insert coffee brewing method into db: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: brewing_method: transaction failed: %w", err)
	}
	return nil
}

func (s *SQLiteDB) insertCoffee(ctx context.Context, coffee coffee) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO coffees(name, roaster, region, variety, method, decaf)
			VALUES (:name, :roaster, :region, :variety, :method, :decaf)
		`,
			sql.Named("name", coffee.name),
			sql.Named("roaster", coffee.roaster),
			sql.Named("region", coffee.region),
			sql.Named("variety", coffee.variety),
			sql.Named("method", coffee.method),
			sql.Named("decaf", coffee.decaf),
		); err != nil {
			return fmt.Errorf("buna: coffee: failed to insert coffee into db: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE coffees
			SET roaster = NULLIF(roaster, ""),
				region = NULLIF(region, ""),
				variety = NULLIF(variety, ""),
				method = NULLIF(method, "")
			WHERE name = :name
		`,
			sql.Named("name", coffee.name),
		); err != nil {
			return fmt.Errorf("buna: coffee: failed to set null values: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: coffee: insertCoffee transaction failed: %w", err)
	}
	return nil
}

func (s *SQLiteDB) insertCoffeePurchase(ctx context.Context, coffeePurchase coffeePurchase) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var coffeeID int
		if err := tx.QueryRowContext(ctx, `
			SELECT id
			FROM coffees
			WHERE name = :coffeeName AND (roaster = :coffeeRoaster OR (:coffeeRoaster = "" AND roaster IS NULL))
		`,
			sql.Named("coffeeName", coffeePurchase.coffeeName),
			sql.Named("coffeeRoaster", coffeePurchase.coffeeRoaster),
		).Scan(&coffeeID); err != nil {
			fmt.Println("Unable to link the purchased coffee to an existing coffee. Please create a new coffee first and then try again.")
			return nil
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO purchases(coffee_id, bought_date, roast_date)
			VALUES (:coffeeID, :boughtDate, :roastDate)
		`,
			sql.Named("coffeeID", coffeeID),
			sql.Named("boughtDate", coffeePurchase.boughtDate),
			sql.Named("roastDate", coffeePurchase.roastDate),
		); err != nil {
			return fmt.Errorf("buna: coffee_purchase: failed to insert coffee purchase into db: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE purchases
			SET roast_date = NULLIF(roast_date, "0-00-00")
			WHERE coffee_id = :coffeeID AND bought_date = :boughtDate
		`,
			sql.Named("coffeeID", coffeeID),
			sql.Named("boughtDate", coffeePurchase.boughtDate),
		); err != nil {
			return fmt.Errorf("buna: coffee_purchase: failed to set null values: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: coffee_purchase: transaction failed: %w", err)
	}
	return nil
}

func (s *SQLiteDB) insertCupping(ctx context.Context, cupping cupping) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO cuppings(date, duration_min, notes)
			VALUES (:cuppingDate, :cuppingDurationMin, :cuppingNotes)
		`,
			sql.Named("cuppingDate", cupping.date),
			sql.Named("cuppingDurationMin", cupping.durationMin),
			sql.Named("cuppingNotes", cupping.notes),
		)
		if err != nil {
			return fmt.Errorf("buna: cupping: failed to insert cupping into db: %w", err)
		}

		cuppingID, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("buna: cupping: failed to get cupping id: %w", err)
		}

		for _, cuppedCoffee := range cupping.cuppedCoffees {
			coffeeID, err := s.getCoffeeIDByNameRoaster(ctx, cuppedCoffee.name, cuppedCoffee.roaster)
			if err != nil {
				return fmt.Errorf("buna: cupping: failed to retrieve coffee id from db: %w", err)
			}

			if _, err := tx.ExecContext(ctx, `
				INSERT INTO cupped_coffees(cupping_id, coffee_id, rank, notes)
				VALUES (:cuppingID, :coffeeID, :coffeeRank, :coffeeNotes)
			`,
				sql.Named("cuppingID", cuppingID),
				sql.Named("coffeeID", coffeeID),
				sql.Named("coffeeRank", cuppedCoffee.rank),
				sql.Named("coffeeNotes", cuppedCoffee.notes),
			); err != nil {
				return fmt.Errorf("buna: cupping: failed to insert cupped coffee into db: %w", err)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: cupping: insert cupping transaction failed: %w", err)
	}
	return nil
}

func (s *SQLiteDB) insertGrinder(ctx context.Context, grinder grinder) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO grinders(name, company, max_grind_setting)
			VALUES (:name, :company, :maxGrindSetting)
		`,
			sql.Named("name", grinder.name),
			sql.Named("company", grinder.company),
			sql.Named("maxGrindSetting", grinder.maxGrindSetting),
		); err != nil {
			return fmt.Errorf("buna: grinder: failed to insert coffee grinder into db: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE grinders
			SET company = NULLIF(company, ""),
				max_grind_setting = NULLIF(max_grind_setting, 0)
			WHERE name = :name
		`,
			sql.Named("name", grinder.name),
		); err != nil {
			return fmt.Errorf("buna: grinder: failed to set null values: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: grinder: transaction failed: %w", err)
	}
	return nil
}

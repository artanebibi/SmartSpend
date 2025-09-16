package repository

import (
	"SmartSpend/internal/database"
	"context"
	"database/sql"
	"time"
)

type IStatisticsRepository interface {
	FindTotalIncomeAndExpense(userId string, from time.Time, to time.Time) (float32, float32, error)
	FindPercentageSpentPerCategory(userId string, from time.Time, to time.Time) (map[string]float32, float32, float32, error)
	FindTotalSpentPerMonth(userId string, from time.Time, to time.Time) (map[int32]float32, error)
	FindAverage(userId string, from time.Time, to time.Time) (float32, float32, error)
}

type databaseStatisticsRepository struct {
	db *sql.DB
}

func NewStatisticsRepository(s database.Service) IStatisticsRepository {
	return &databaseStatisticsRepository{
		db: s.DB(),
	}
}
func (r *databaseStatisticsRepository) FindTotalIncomeAndExpense(userId string, from time.Time, to time.Time) (float32, float32, error) {
	row := r.db.QueryRow(`
		WITH total_expense AS (
			SELECT COALESCE(SUM(price), 0) AS total
			FROM transactions
			WHERE owner_id = $1 AND type = 'Expense' AND date_made BETWEEN $2 AND $3
		),
		total_income AS (
			SELECT COALESCE(SUM(price), 0) AS total
			FROM transactions
			WHERE owner_id = $1 AND type = 'Income' AND date_made BETWEEN $2 AND $3
		)
		SELECT 
			(SELECT total FROM total_expense),
			(SELECT total FROM total_income)
	`, userId, from, to)

	var totalExpense float32
	var totalIncome float32
	if err := row.Scan(&totalExpense, &totalIncome); err != nil {
		return 0, 0, err
	}
	return totalExpense, totalIncome, nil

}

func (r *databaseStatisticsRepository) FindPercentageSpentPerCategory(userId string, from time.Time, to time.Time) (map[string]float32, float32, float32, error) {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead, // needed so all the reads see the same snapshot of the data when the transaction is being ran.
		ReadOnly:  true,
	})
	if err != nil {
		return nil, 0, 0, err
	}
	defer tx.Rollback()

	query := `
		WITH total_expense AS (
			SELECT owner_id, SUM(price) AS total_expense
			FROM transactions
			WHERE owner_id = $1
			  AND type = 'Expense'
			  AND date_made BETWEEN $2 AND $3
			GROUP BY owner_id
		),
		total_income AS (
			SELECT owner_id, SUM(price) AS total_income
			FROM transactions
			WHERE owner_id = $1
			  AND type = 'Income'
			  AND date_made BETWEEN $2 AND $3
			GROUP BY owner_id
		)
		SELECT 
			t.owner_id,
			c.name,
			SUM(t.price) AS total_per_category,
			(SUM(t.price) / COALESCE(te.total_expense, 1)) * 100.0 AS percentage_per_category,
			COALESCE(te.total_expense, 0) AS total_expense,
			COALESCE(ti.total_income, 0) AS total_income
		FROM transactions t
		LEFT JOIN total_expense te
			ON t.owner_id = te.owner_id
		JOIN categories c
			ON c.id = t.category_id
		LEFT JOIN total_income ti
			ON t.owner_id = ti.owner_id
		WHERE 
			t.owner_id = $1
			AND t.type = 'Expense'
			AND t.date_made BETWEEN $2 AND $3
		GROUP BY t.owner_id, c.name, te.total_expense, ti.total_income;
`

	rows, err := tx.Query(query, userId, from, to)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	percentages := make(map[string]float32)
	var totalExpense float32
	var totalIncome float32

	for rows.Next() {
		var ownerId string
		var category string
		var total float32
		var percentage float32
		var totalUserExpense float32
		var totalUserIncome float32

		if err := rows.Scan(&ownerId, &category, &total, &percentage, &totalUserExpense, &totalUserIncome); err != nil {
			return nil, 0, 0, err
		}
		percentages[category] = percentage
		totalExpense = totalUserExpense
		totalIncome = totalUserIncome
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, 0, err
	}

	return percentages, totalExpense, totalIncome, nil
}

func (r *databaseStatisticsRepository) FindTotalSpentPerMonth(userId string, from time.Time, to time.Time) (map[int32]float32, error) {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT 
			EXTRACT(month FROM date_made) AS month,
			SUM(price) AS value_spent
		FROM transactions
		WHERE owner_id = $1
		  AND type = 'Expense'
		  AND date_made BETWEEN $2 AND $3
		GROUP BY month
		ORDER BY month;
	`

	rows, err := tx.Query(query, userId, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spentPerMonth := make(map[int32]float32)

	for rows.Next() {
		var month int32
		var valueSpent float32

		if err := rows.Scan(&month, &valueSpent); err != nil {
			return nil, err
		}

		spentPerMonth[month] = valueSpent
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return spentPerMonth, nil
}

func (r *databaseStatisticsRepository) FindAverage(userId string, from time.Time, to time.Time) (float32, float32, error) {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	query := `
	WITH average_expense AS (
		SELECT owner_id, COALESCE(AVG(price), 0) AS average_expense
		FROM transactions
		WHERE owner_id = $1
		  AND type = 'Expense'
		  AND date_made BETWEEN $2 AND $3
		GROUP BY owner_id
	),
	average_income AS (
		SELECT owner_id, COALESCE(AVG(price), 0) AS average_income
		FROM transactions
		WHERE owner_id = $1
		  AND type = 'Income'
		  AND date_made BETWEEN $2 AND $3
		GROUP BY owner_id
	),
	base AS (
		SELECT DISTINCT owner_id
		FROM transactions
		WHERE owner_id = $1
	)
	SELECT 
		b.owner_id,
		COALESCE(ae.average_expense, 0) AS average_expense,
		COALESCE(ai.average_income, 0) AS average_income
	FROM base b
	LEFT JOIN average_expense ae ON b.owner_id = ae.owner_id
	LEFT JOIN average_income ai ON b.owner_id = ai.owner_id;
	`

	row := tx.QueryRow(query, userId, from, to)

	var ownerId string
	var averageExpense float32
	var averageIncome float32

	if err := row.Scan(&ownerId, &averageExpense, &averageIncome); err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return averageExpense, averageIncome, nil
}

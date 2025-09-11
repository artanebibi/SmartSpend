package repository

import (
	"SmartSpend/internal/database"
	"context"
	"database/sql"
	"time"
)

type IStatisticsRepository interface {
	FindPercentageSpentPerCategory(userId string, from time.Time, to time.Time) (map[string]float32, float32, float32, error)
}

type databaseStatisticsRepository struct {
	db *sql.DB
}

func NewStatisticsRepository(s database.Service) IStatisticsRepository {
	return &databaseStatisticsRepository{
		db: s.DB(),
	}
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

	query :=
		`
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
				(SUM(t.price) / te.total_expense) * 100.0 AS percentage_per_category,
				te.total_expense,
				ti.total_income
			FROM transactions t
			JOIN total_expense te
				ON t.owner_id = te.owner_id
			JOIN categories c
				ON c.id = t.category_id
			JOIN total_income ti
				ON t.owner_id = ti.owner_id
			WHERE 
			    t.owner_id = $1 AND t.type = 'Expense' AND t.date_made BETWEEN $2 AND $3
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

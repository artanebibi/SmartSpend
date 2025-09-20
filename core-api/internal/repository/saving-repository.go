package repository

import (
	"SmartSpend/internal/database"
	"SmartSpend/internal/domain/model"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

type ISavingRepository interface {
	FindAll(userId string, from time.Time, to time.Time) []model.Saving
	FindById(id int64, userId string) (*model.Saving, error)
	Save(saving model.Saving) error
	Update(saving model.Saving, id int64) error
	Delete(id int64, userId string) error
}

type databaseSavingRepository struct {
	db *sql.DB
}

func (d *databaseSavingRepository) FindAll(userId string, from time.Time, to time.Time) []model.Saving {
	rows, err := d.db.Query(`
		SELECT id, owner_id, amount, "from", "to"
		FROM savings
		WHERE owner_id = $1
		  AND date_made >= $2
		  AND date_made <= $3
		ORDER BY date_made ASC
	`, userId, from, to)

	if err != nil {
		return nil
	}
	defer rows.Close()

	var savings []model.Saving
	for rows.Next() {
		var t model.Saving
		if err := rows.Scan(&t.ID, &t.OwnerId, &t.Amount, &t.From, &t.To); err != nil {
			log.Println(err)
			continue
		}
		savings = append(savings, t)
	}
	return savings
}

func NewSavingRepository(s database.Service) ISavingRepository {
	return &databaseSavingRepository{
		db: s.DB(),
	}
}

func (d *databaseSavingRepository) FindById(id int64, userId string) (*model.Saving, error) {
	row := d.db.QueryRow(
		`SELECT * FROM savings WHERE id = $1 and owner_id = $2`,
		id, userId,
	)

	var saving model.Saving
	err := row.Scan(
		&saving.ID,
		&saving.OwnerId,
		&saving.Amount,
		&saving.From,
		&saving.To,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("saving not found")
		}
		return nil, fmt.Errorf("failed to scan saving: %v", err)
	}

	return &saving, nil
}

func (d *databaseSavingRepository) Save(saving model.Saving) error {
	log.Println("Saving service", saving)
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
								INSERT INTO savings (id, owner_id, amount, from, to)
    					        VALUES ($1, $2, $3, $4, $5)
    					        `,
		saving.ID, saving.OwnerId, saving.Amount, saving.From, saving.To)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *databaseSavingRepository) Update(saving model.Saving, id int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(
		`UPDATE savings
         SET amount = $1,
             from = $2,
             to = $3
         WHERE id = $4`,
		saving.Amount, saving.From, saving.To, id)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *databaseSavingRepository) Delete(id int64, userId string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	t, err := d.FindById(id, userId)

	if t == nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM savings WHERE id = $1 AND owner_id = $2", id, userId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

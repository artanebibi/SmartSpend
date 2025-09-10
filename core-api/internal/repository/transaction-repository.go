package repository

import (
	"SmartSpend/internal/database"
	"SmartSpend/internal/domain/enum"
	"SmartSpend/internal/domain/model"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type ITransactionRepository interface {
	FindAll(userId string) []model.Transaction
	FindById(id int64, userId string) (*model.Transaction, error)
	Save(transaction model.Transaction) error
	Update(transaction model.Transaction, id int64) error
	Delete(id int64, userId string) error
}

type databaseTransactionRepository struct {
	db *sql.DB
}

func (d *databaseTransactionRepository) Save(transaction model.Transaction) error {
	log.Println("Saving transaction:", transaction.Title)

	if transaction.Type == enum.Income {
		tx, err := d.db.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
								INSERT INTO transactions (title, price, date_made, owner_id, category_id, type)
    					        VALUES ($1, $2, $3, $4, NULL, $5)
    					        `,
			transaction.Title, transaction.Price, transaction.DateMade, transaction.OwnerId, transaction.Type)
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec(`
    							UPDATE users
							    SET balance = balance + $1
    							WHERE id = $2
								`,
			transaction.Price, transaction.OwnerId)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
		return err
	} else { // Expense
		tx, err := d.db.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
								INSERT INTO transactions (title, price, date_made, owner_id, category_id, type)
    					        VALUES ($1, $2, $3, $4, $5, $6)
    					        `,
			transaction.Title, transaction.Price, transaction.DateMade, transaction.OwnerId, transaction.CategoryId, transaction.Type)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = tx.Exec(`
    							UPDATE users
							    SET balance = balance - $1
    							WHERE id = $2
								`,
			transaction.Price, transaction.OwnerId)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
		return err
	}
}

func (d *databaseTransactionRepository) Update(transaction model.Transaction, id int64) error {
	log.Println("Updating transaction:", transaction)

	var categoryId interface{}
	if transaction.CategoryId == nil {
		categoryId = nil
	} else {
		categoryId = *transaction.CategoryId
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var oldPrice float32
	var oldType string
	var ownerId string
	err = tx.QueryRow(`SELECT price, type, owner_id FROM transactions WHERE id = $1`, id).Scan(&oldPrice, &oldType, &ownerId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`UPDATE transactions
         SET title = $1,
             price = $2,
             date_made = $3,
             category_id = $4,
             "type" = $5
         WHERE id = $6`,
		transaction.Title, transaction.Price, transaction.DateMade,
		categoryId, transaction.Type, id,
	)
	if err != nil {
		return err
	}

	var balanceAdjustment float32
	if oldType == "Expense" {
		balanceAdjustment += oldPrice
	} else if oldType == "Income" {
		balanceAdjustment -= oldPrice
	}

	if transaction.Type == "Expense" {
		balanceAdjustment -= transaction.Price
	} else if transaction.Type == "Income" {
		balanceAdjustment += transaction.Price
	}

	_, err = tx.Exec(
		`UPDATE users SET balance = balance + $1 WHERE id = $2`,
		balanceAdjustment, ownerId,
	)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func NewTransactionRepository(s database.Service) ITransactionRepository {
	return &databaseTransactionRepository{
		db: s.DB(),
	}
}

func (d *databaseTransactionRepository) FindAll(userId string) []model.Transaction {
	rows, err := d.db.Query(`SELECT * FROM transactions WHERE owner_id = $1`, userId)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(&t.ID, &t.Title, &t.Price, &t.DateMade, &t.OwnerId, &t.CategoryId, &t.Type); err != nil {
			log.Println(err)
			continue
		}
		transactions = append(transactions, t)
	}
	return transactions
}

func (d *databaseTransactionRepository) FindById(id int64, userId string) (*model.Transaction, error) {
	row := d.db.QueryRow(
		`SELECT * FROM transactions WHERE id = $1 and owner_id = $2`,
		id, userId,
	)

	var transaction model.Transaction
	err := row.Scan(
		&transaction.ID,
		&transaction.Title,
		&transaction.Price,
		&transaction.DateMade,
		&transaction.OwnerId,
		&transaction.CategoryId,
		&transaction.Type,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to scan transaction: %v", err)
	}

	return &transaction, nil
}

func (d *databaseTransactionRepository) Delete(id int64, userId string) error {
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

	_, err = tx.Exec("DELETE FROM transactions WHERE id = $1 AND owner_id = $2", id, userId)
	if err != nil {
		return err
	}

	var balanceAdjustment float32
	if t.Type == "Expense" {
		balanceAdjustment += t.Price
	} else if t.Type == "Income" {
		balanceAdjustment = -t.Price
	}

	_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", balanceAdjustment, userId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

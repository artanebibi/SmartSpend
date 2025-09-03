package repository

import (
	"SmartSpend/internal/database"
	"SmartSpend/internal/domain/enum"
	"SmartSpend/internal/domain/model"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type ITransactionRepository interface {
	FindAll(userId uuid.UUID) []model.Transaction
	FindById(id int64, userId uuid.UUID) (*model.Transaction, error)
	Save(transaction model.Transaction) error
	Update(transaction model.Transaction, id int64) error
	Delete(id int64, userId uuid.UUID) error
}

type databaseTransactionRepository struct {
	db *sql.DB
}

func (d *databaseTransactionRepository) Save(transaction model.Transaction) error {
	log.Println("Saving transaction:", transaction.Title)

	if transaction.Type == enum.Income {
		_, err := d.db.Exec(
			`INSERT INTO transactions (title, price, date_made, owner_id, category_id, "type")
         VALUES ($1, $2, $3, $4, null, $5)`,
			transaction.Title,
			transaction.Price,
			transaction.DateMade,
			transaction.OwnerId,
			transaction.Type,
		)
		return err
	}

	_, err := d.db.Exec(
		`INSERT INTO transactions (title, price, date_made, owner_id, category_id, "type")
         VALUES ($1, $2, $3, $4, $5, $6)`,
		transaction.Title,
		transaction.Price,
		transaction.DateMade,
		transaction.OwnerId,
		transaction.CategoryId,
		transaction.Type,
	)

	return err
}

func (d *databaseTransactionRepository) Update(transaction model.Transaction, id int64) error {
	log.Println("Updating transaction :", transaction)

	var categoryId interface{}
	if transaction.CategoryId == nil {
		categoryId = nil
	} else {
		categoryId = *transaction.CategoryId
	}

	_, err := d.db.Exec(
		`UPDATE transactions
		 SET title = $1,
		     price = $2,
		     date_made = $3,
		     owner_id = $4,
		     category_id = $5,
		     "type" = $6
		 WHERE id = $7`,
		transaction.Title,
		transaction.Price,
		transaction.DateMade,
		transaction.OwnerId,
		categoryId,
		transaction.Type,
		id,
	)

	return err
}

func NewTransactionRepository(s database.Service) ITransactionRepository {
	return &databaseTransactionRepository{
		db: s.DB(),
	}
}

func (d *databaseTransactionRepository) FindAll(userId uuid.UUID) []model.Transaction {
	rows, err := d.db.Query("SELECT * FROM transactions WHERE owner_id = $1", userId)
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

func (d *databaseTransactionRepository) FindById(id int64, userId uuid.UUID) (*model.Transaction, error) {
	row := d.db.QueryRow(
		"SELECT * FROM transactions WHERE id = $1 and owner_id = $2",
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

func (d *databaseTransactionRepository) Delete(id int64, userId uuid.UUID) error {
	row := d.db.QueryRow(
		"DELETE FROM transactions WHERE id = $1 and owner_id = $2",
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
			return fmt.Errorf("transaction not found")
		}
		return fmt.Errorf("failed to scan transaction: %v", err)
	}

	return nil
}

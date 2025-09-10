package repository

import (
	"SmartSpend/internal/database"
	"SmartSpend/internal/domain/model"
	"database/sql"
	"log"
)

type ICategoryRepository interface {
	FindAll() []model.Category
}

type databaseCategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(s database.Service) ICategoryRepository {
	return &databaseCategoryRepository{
		db: s.DB(),
	}
}

func (d *databaseCategoryRepository) FindAll() []model.Category {
	rows, err := d.db.Query("SELECT * FROM categories")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var t model.Category
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			log.Println(err)
			continue
		}
		categories = append(categories, t)
	}
	return categories
}

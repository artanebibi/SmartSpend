package repository

import (
	"SmartSpend/internal/database"
	"SmartSpend/internal/domain/model"
	"database/sql"
	"fmt"
	"log"
)

type IUserRepository interface {
	FindAll() []model.User
	FindById(id string) (*model.User, error)
	FindByGoogleEmail(email string) *model.User
	FindByAppleEmail(email string) *model.User
	Save(user model.User) error
	Update(user model.User) error
	Delete(id string) error
}

type databaseUserRepository struct {
	db *sql.DB
}

func NewUserRepository(s database.Service) IUserRepository {
	return &databaseUserRepository{
		db: s.DB(),
	}
}

func (d *databaseUserRepository) FindAll() []model.User {
	rows, err := d.db.Query("SELECT id, first_name, last_name, google_email FROM users")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.GoogleEmail); err != nil {
			log.Println(err)
			continue
		}
		users = append(users, u)
	}
	return users
}

func (d *databaseUserRepository) FindById(id string) (*model.User, error) {
	row := d.db.QueryRow(
		"SELECT * FROM users WHERE id = $1",
		id,
	)

	var user model.User
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.GoogleEmail,
		&user.AppleEmail,
		&user.RefreshToken,
		&user.RefreshTokenExpiryDate,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.Balance,
		&user.MonthlySavingGoal,
		&user.PreferredCurrency,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user: %v", err)
	}

	return &user, nil
}

func (d *databaseUserRepository) FindByGoogleEmail(email string) *model.User {
	row := d.db.QueryRow(
		"SELECT * FROM users WHERE google_email = $1",
		email,
	)

	var user model.User
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.GoogleEmail,
		&user.AppleEmail,
		&user.RefreshToken,
		&user.RefreshTokenExpiryDate,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.Balance,
		&user.MonthlySavingGoal,
		&user.PreferredCurrency,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("user with google_email=%s not found", email)
			return nil
		}
		log.Printf("failed to scan user by email: %v", err)
		return nil
	}

	return &user
}

func (d *databaseUserRepository) FindByAppleEmail(email string) *model.User {
	row := d.db.QueryRow(
		"SELECT * FROM users WHERE apple_email = $1",
		email,
	)

	var user model.User
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.GoogleEmail,
		&user.AppleEmail,
		&user.RefreshToken,
		&user.RefreshTokenExpiryDate,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.Balance,
		&user.MonthlySavingGoal,
		&user.PreferredCurrency,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("user with google_email=%s not found", email)
			return nil
		}
		log.Printf("failed to scan user by email: %v", err)
		return nil
	}

	return &user
}

func (d *databaseUserRepository) Save(user model.User) error {
	log.Println("Saving user:", user.FirstName)
	_, err := d.db.Exec(
		"INSERT INTO users (id,first_name,last_name,username, google_email,apple_email,refresh_token,refresh_token_expiry_date,avatar_url,created_at,balance,monthly_saving_goal,preferred_currency) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
		user.ID, user.FirstName, user.LastName, user.Username, user.GoogleEmail, user.AppleEmail, user.RefreshToken, user.RefreshTokenExpiryDate, user.AvatarURL, user.CreatedAt, user.Balance, user.MonthlySavingGoal, user.PreferredCurrency,
	)
	return err
}

func (d *databaseUserRepository) Update(user model.User) error {
	_, err := d.db.Exec(
		`
		UPDATE users
		
		 SET first_name = $1,
		     last_name = $2,
		     username = $3,
		     google_email = $4,
		     apple_email = $5,
		     refresh_token = $6,
		     refresh_token_expiry_date = $7,
		     avatar_url = $8,
		     balance = $9,
		     monthly_saving_goal = $10,
		     preferred_currency = $12
		
	    WHERE id = $11`,
		user.FirstName,
		user.LastName,
		user.Username,
		user.GoogleEmail,
		user.AppleEmail,
		user.RefreshToken,
		user.RefreshTokenExpiryDate,
		user.AvatarURL,
		user.Balance,
		user.MonthlySavingGoal,
		user.ID,
		user.PreferredCurrency,
	)
	if err != nil {
		log.Printf("failed to update user %s: %v", user.ID, err)
		return err
	}
	return nil
}

func (d *databaseUserRepository) Delete(id string) error {
	d.db.Exec("DELETE FROM users WHERE id = $1", id)
	return nil
}

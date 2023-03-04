package dbrepo

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Atul-Ranjan12/tourism/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// This package contains all the functionalities and queries of the database
// Currently this application is using the postgres mailing server

// InsertNewUser adds a new registration to the database
func (m *PostgresDBRepo) InsertNewUser(reg models.UserRegistration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (first_name, last_name, email, 
			password, age, gender, access_level, phone_number, mail_verification_code, created_at, updated_at)
		VALUES  ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := m.DB.ExecContext(ctx, query,
		reg.FirstName,
		reg.LastName,
		reg.Email,
		reg.HashedPassword,
		reg.Age,
		reg.Gender,
		1,
		reg.PhoneNumber,
		reg.VerificationCode,
		reg.CreatedAt,
		reg.UpdatedAt,
	)
	if err != nil {
		log.Println("Error eexecuting the query")
		return err
	}

	return nil
}

func (m *PostgresDBRepo) UserExists(email string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	SELECT COUNT(*) FROM users WHERE email = $1
	`
	var numRows int
	row := m.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&numRows)
	if err != nil {
		log.Println("ERROR: Error executing query to check if the user exists")
		return false, err
	}

	if numRows == 0 {
		return false, nil
	}
	return true, nil
}

func (m *PostgresDBRepo) CheckTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select * from merchants
	`

	_, err := m.DB.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	log.Println("Succesful query execution")
	return nil
}

// Function to Authenticate the User for a sign in
func (m *PostgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "SELECT id, password from users WHERE email = $1", email)

	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	// Compare password with my password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		log.Println("Invalid Password")
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		log.Println("Unexpected Error")
		return 0, "", err
	}

	return id, hashedPassword, nil
}

func (m *PostgresDBRepo) FindUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, first_name, last_name, email, phone_number, user_is_verified
		FROM users
		WHERE id = $1
	`
	var user models.User

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Phone,
		&user.IsVerified,
	)
	if err != nil {
		log.Println("Probable error in query execution: FindUserBID")
		return user, err
	}

	return user, nil
}

func (m *PostgresDBRepo) IncrementVerification(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE users
		SET user_is_verified = $1
		WHERE id = $2
	`

	_, err := m.DB.ExecContext(ctx, query, user.IsVerified+1, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m *PostgresDBRepo) GetVerificationCode(user models.User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var verificationCode int

	query := `
		SELECT mail_verification_code 
		FROM users
		WHERE id = $1
	`

	row := m.DB.QueryRowContext(ctx, query, user.ID)
	err := row.Scan(&verificationCode)
	if err != nil {
		return 0, err
	}

	return verificationCode, err
}

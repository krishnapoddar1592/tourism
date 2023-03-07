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

// Function to find the user by ID
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

	log.Println("Increment Verification Called :: Incrementing to: ", user.IsVerified+1)

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

// Funciton to get the verification code from the user
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

// Function to add the merchant address to the database
func (m *PostgresDBRepo) AddMerchantAddress(address models.MerchantAddress) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO merchant_address (city, state, country, address_line_1, address_line_2, created_at, updated_at, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := m.DB.QueryContext(ctx, query,
		address.City,
		address.State,
		address.Country,
		address.AddressLine1,
		address.AddressLine2,
		address.CreatedAt,
		address.UpdatedAt,
		address.UserID,
	)
	if err != nil {
		log.Println("Error executing query: Inserting into merchant_address")
		return err
	}

	return nil
}

func (m *PostgresDBRepo) AddMerchantDocuments(docs models.MerchantDocument) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO merchant_documents (document_link, document_id, image, created_at, updated_at, user_id)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`
	var newID int

	err := m.DB.QueryRowContext(ctx, query,
		docs.DocumentLink,
		docs.DocumentID,
		docs.ImageFile,
		docs.CreatedAt,
		docs.UpdatedAt,
		docs.UserID,
	).Scan(&newID)

	if err != nil {
		log.Println("Error executing query : Inserting into merchant_documents")
		return 0, err
	}
	return newID, nil
}

// Funciton adds user and document references to the merchants table
func (m *PostgresDBRepo) AddMerchant(mer models.MerchantData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO merchants (user_id, address_id, document_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := m.DB.QueryContext(ctx, query,
		mer.UserID,
		mer.AddressID,
		mer.DocumentID,
		mer.CreatedAt,
		mer.UpdatedAt,
	)
	if err != nil {
		log.Println("Error executing query : Inserting into merchants tablel")
		return err
	}
	return nil
}

func (m *PostgresDBRepo) GetAddressIDFromUser(userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id FROM merchant_address WHERE user_id = $1
	`
	var addressID int
	row := m.DB.QueryRowContext(ctx, query, userID)
	err := row.Scan(&addressID)
	if err != nil {
		return 0, err
	}
	return addressID, err
}

// Function to get the userid form the merchatn id
func (m *PostgresDBRepo) GetMerchantIDFromUserID(userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id FROM merchants WHERE user_id = $1`

	var merchantID int
	row := m.DB.QueryRowContext(ctx, query, userID)
	err := row.Scan(&merchantID)
	if err != nil {
		return 0, err
	}

	return merchantID, nil
}

// Add teh bus details to the server
func (m *PostgresDBRepo) AddBusToDatabase(bus models.AddBusData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO bus (bus_name, bus_source, bus_destination, bus_model, bus_no_plate, num_seats, office_pan, office_address, 
						merchant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := m.DB.QueryContext(ctx, query,
		bus.BusName,
		bus.BusStart,
		bus.BusEnd,
		bus.BusModel,
		bus.BusNumPlate,
		bus.BusNumSeats,
		bus.BusPAN,
		bus.BusAddress,
		bus.MerchantID,
		bus.CreatedAt,
		bus.UpdatedAt,
	)
	if err != nil {
		log.Println("Error executing query: ", err)
		return err
	}

	return nil
}

func (m *PostgresDBRepo) GetAllBus(merchantID int) ([]models.AddBusData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT bus_name, bus_source, bus_destination, bus_model, bus_no_plate, num_seats, office_pan, office_address, 
			   merchant_id, created_at, updated_at
		FROM bus
		WHERE merchant_id = $1
	`
	var busses []models.AddBusData

	rows, err := m.DB.QueryContext(ctx, query, merchantID)
	if err != nil {
		log.Println("Could not execute query: GetAllBus ", err)
	}
	defer rows.Close()

	for rows.Next() {
		var i models.AddBusData
		err := rows.Scan(
			&i.BusName,
			&i.BusStart,
			&i.BusEnd,
			&i.BusModel,
			&i.BusNumPlate,
			&i.BusNumSeats,
			&i.BusPAN,
			&i.BusAddress,
			&i.MerchantID,
			&i.CreatedAt,
			&i.UpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning the rows into variables")
			return busses, err
		}

		busses = append(busses, i)
	}
	if err = rows.Err(); err != nil {
		return busses, err
	}
	return busses, nil
}

package models

import "time"

type UserRegistration struct {
	FirstName        string
	LastName         string
	Age              string
	Gender           string
	Email            string
	HashedPassword   string
	PhoneNumber      string
	VerificationCode int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type User struct {
	ID         int
	FirstName  string
	LastName   string
	Email      string
	Phone      string
	IsVerified int
}

// Structure for the mail data
type ConfirmationMailData struct {
	To      string
	From    string
	Subject string
	Content string
}

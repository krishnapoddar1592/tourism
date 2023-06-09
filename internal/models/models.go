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

type MerchantAddress struct {
	City         string
	State        string
	Country      string
	AddressLine1 string
	AddressLine2 string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UserID       int
}

type MerchantDocument struct {
	DocumentID   string
	DocumentLink string
	ImageFile    []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UserID       int
}

type MerchantData struct {
	UserID     int
	AddressID  int
	DocumentID int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Type to add the bus form
type AddBusData struct {
	BusID       int
	MerchantID  int
	BusName     string
	BusModel    string
	BusAddress  string
	BusStart    string
	BusEnd      string
	BusNumSeats int
	BusNumPlate string
	BusPAN      string
	Price       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Bus Reservation Model
type BusReservationData struct {
	ReservationID   int
	BusID           int
	FirstName       string
	LastName        string
	ReservationDate time.Time
	NumPassengers   int
	From            string
	Stop            string
	PhoneNumber     string
	Email           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Bus             AddBusData
}

// Model for the Hotel/ Hotel Room
type HotelRoom struct {
	HotelID              int
	MerchantID           int
	HotelName            string
	HotelRoomName        string
	HotelType            string
	HotelAddress         string
	HotelPAN             string
	HotelNumRooms        int
	HotelPhone1          string
	HotelPhone2          string
	HotelRoomDescription string
	Price                int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

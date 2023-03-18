package repository

import "github.com/Atul-Ranjan12/tourism/internal/models"

type DatabaseRepo interface {
	InsertNewUser(reg models.UserRegistration) error
	CheckTable() error
	Authenticate(email, testPassword string) (int, string, error)
	FindUserByID(id int) (models.User, error)
	UserExists(email string) (bool, error)
	GetVerificationCode(user models.User) (int, error)
	IncrementVerification(user models.User) error
	AddMerchantAddress(address models.MerchantAddress) error
	AddMerchantDocuments(docs models.MerchantDocument) (int, error)
	AddMerchant(mer models.MerchantData) error
	GetAddressIDFromUser(userID int) (int, error)
	GetMerchantIDFromUserID(userID int) (int, error)
	AddBusToDatabase(bus models.AddBusData) error
	GetAllBus(merchantID int) ([]models.AddBusData, error)
	GetBusByID(busID int) (models.AddBusData, error)
	UpdateBusInfo(busID int, i models.AddBusData) error
	DeleteBusByID(busID int) error
	MakeBusReservation(busRes models.BusReservationData) error
}

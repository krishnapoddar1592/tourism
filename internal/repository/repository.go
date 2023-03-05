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
}

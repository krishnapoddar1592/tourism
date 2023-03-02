package forms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
	"golang.org/x/crypto/bcrypt"
)

// Defining a custom form struct
type Form struct {
	url.Values
	Errors errors
}

// Initializes a form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

func (f *Form) Has(key string) bool {
	x := f.Get(key)
	if x == "" {
		f.Errors.Add(key, "This field cannot be empty")
		return false
	}
	return true
}

// Checks all of the requuired fields in a generalized way
func (f *Form) Required(keys ...string) {
	for _, key := range keys {
		value := f.Get(key)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(key, "This field cannot be blank")
		}
	}
}

// Checks for minimum length
func (f *Form) MinLength(key string, length int, r *http.Request) bool {
	value := f.Get(key)
	if len(value) < length {
		f.Errors.Add(key, fmt.Sprintf("Minimum number of length shoulld be: %d", length))
		return false
	}
	return true
}

// Checks if Email is valid
func (f *Form) IsEmailValid(key string) {
	if !govalidator.IsEmail(f.Get(key)) {
		f.Errors.Add(key, "Please enter a valid Email address")
	}
}

// Returns true if there are no errors
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// Validate if the function is valid
func (f *Form) IsPasswordValid(password, confirmPassword string) {
	hash1, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	hash2, _ := bcrypt.GenerateFromPassword([]byte(confirmPassword), 12)

	if string(hash1) != string(hash2) {
		f.Errors.Add(confirmPassword, "Password not matching")
		return
	} else {
		return
	}
}

// Check if user has checked the box
func (f *Form) HasUserAccepted(key string) {
	if f.Get(key) == "agreed" {
		return
	} else {
		f.Errors.Add(key, "Have to agree to the terms of service")
	}
}

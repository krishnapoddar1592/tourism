package forms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
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

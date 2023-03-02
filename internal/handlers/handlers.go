package handlers

import (
	"log"
	"net/http"

	"github.com/Atul-Ranjan12/tourism/internal/config"
	"github.com/Atul-Ranjan12/tourism/internal/driver"
	"github.com/Atul-Ranjan12/tourism/internal/forms"
	"github.com/Atul-Ranjan12/tourism/internal/models"
	"github.com/Atul-Ranjan12/tourism/internal/render"
	"github.com/Atul-Ranjan12/tourism/internal/repository"
	"github.com/Atul-Ranjan12/tourism/internal/repository/dbrepo"
)

// Initialize the repository for the application
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

var Repo *Repository

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// Function to create a new test Repository
func NewTestingRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{})
}

// Post Login form
func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")

	log.Println(form.Errors)

	if !form.Valid() {
		log.Println("Invalid login attempt")
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

}

// Show Sign up page
func (m *Repository) ShowSignUp(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["registration"] = models.UserRegistration{
		FirstName:   "",
		LastName:    "",
		Email:       "",
		PhoneNumber: "",
		Age:         "",
		Gender:      "",
	}

	render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// Post the sign up form
func (m *Repository) PostSignUp(w http.ResponseWriter, r *http.Request) {
	// Parsing the form to check for errors and form items
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	// Get the Registration
	registration := models.UserRegistration{
		FirstName:   r.Form.Get("firstName"),
		LastName:    r.Form.Get("lastName"),
		Email:       r.Form.Get("email"),
		PhoneNumber: r.Form.Get("phone"),
		Age:         r.Form.Get("age"),
		Gender:      r.Form.Get("gender"),
	}

	form := forms.New(r.PostForm)

	form.Required("firstName", "lastName", "email", "phone", "age", "gender")
	form.IsEmailValid("email")
	// Validate password and confirmpassword
	form.IsPasswordValid("password", "confirmPassword")
	// Check if the user has clicked on the terms of service
	form.HasUserAccepted("agreeTerms")

	if !form.Valid() {
		data := make(map[string]interface{})
		// Add the registration data to the template
		data["registration"] = registration
		render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

}

package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/Atul-Ranjan12/tourism/internal/config"
	"github.com/Atul-Ranjan12/tourism/internal/driver"
	"github.com/Atul-Ranjan12/tourism/internal/forms"
	"github.com/Atul-Ranjan12/tourism/internal/helpers"
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

	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostLogin Posts the login form
func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	// Prevents session attacks
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid Credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// Put user id and logged in status in the session
	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged In Succesfully")

	//TODO:  Add the User Model in the session
	user, err := m.DB.FindUserByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "user_details", user)

	//TODO: Check if the user is a verified user, if yes, display admin-dashboard
	// else display the verification procedure form.

	// Redirect to the merchant dashboard with the id in the url
	if user.IsVerified > 2 {
		// User is verified
		http.Redirect(w, r, fmt.Sprintf("/merchant/%d/dashboard", id), http.StatusSeeOther)
	} else if user.IsVerified == 0 {
		// User is not verified :: initial
		http.Redirect(w, r, fmt.Sprintf("/merchant/%d/verification", id), http.StatusSeeOther)
	} else if user.IsVerified == 1 {
		// User has completed one step of verification
		http.Redirect(w, r, fmt.Sprintf("/merchant/%d/verification-address", id), http.StatusSeeOther)
	} else {
		// User has completed two levels of verification
		http.Redirect(w, r, fmt.Sprintf("/merchant/%d/verification-documents", id), http.StatusSeeOther)
	}
	//TODO: Create Administrative Pages and Dashboards
	//TODO: Data Breach for Password FIX
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
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	render.Template(w, r, "signup.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostSignUp Handles when the form has been posted
func (m *Repository) PostSignUp(w http.ResponseWriter, r *http.Request) {
	// Parsing the form to check for errors and form items
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
		return
	}
	// Validate form things
	form := forms.New(r.PostForm)

	form.Required("firstName", "lastName", "email", "phone", "age", "gender")
	form.IsEmailValid("email")
	// Validate password and confirmpassword
	form.IsPasswordValid("password", "confirmPassword")
	// Check if the user has clicked on the terms of service
	form.HasUserAccepted("agreeTerms")

	// Hash the password if form is valid
	hashPassword, err := forms.HashPassword(r.Form.Get("password"))
	if err != nil {
		log.Println("Unable to hash password")
		return
	}

	// Check if the user already exists in the database
	email := r.Form.Get("email")

	userExists, err := m.DB.UserExists(email)
	if err != nil {
		log.Println("Error executing the query; error message: ", err)
	}
	form.FormValidateUser("email", userExists)

	// Generate a random 4 digit integer and send it via email
	rand.Seed(time.Now().UnixNano())
	verificationCode := rand.Intn(9000) + 1000

	// Get the Registration
	registration := models.UserRegistration{
		FirstName:        r.Form.Get("firstName"),
		LastName:         r.Form.Get("lastName"),
		Email:            email,
		HashedPassword:   hashPassword,
		PhoneNumber:      r.Form.Get("phone"),
		Age:              r.Form.Get("age"),
		Gender:           r.Form.Get("gender"),
		VerificationCode: verificationCode,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

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

	err = m.DB.InsertNewUser(registration)
	if err != nil {
		log.Println("Error adding the new user to the database with error message... ", err)
		return
	}

	//TODO: Send the verification code through email
	mailMessage := fmt.Sprintf(`
		<h2><strong>Email Verification </strong></h2> <br>
		Dear %s, <br>
		The Verification Code for your email address is: <br>
		<h4> %d </h4>
		Please enter this in the Admin dashboard as asked to verify your email address <br>
		<br><br><br>
		Yours Sincerely, <br>
		TourNepal Inc
	`, registration.FirstName, verificationCode)

	msg := models.ConfirmationMailData{
		To:      registration.Email,
		From:    "info@tournepal.com",
		Subject: "Regarding Email Verification",
		Content: mailMessage,
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "flash", "Succesfully Signed Up")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Function to show the administrative things
func (m *Repository) ShowAdminDashboard(w http.ResponseWriter, r *http.Request) {
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	stringMap := make(map[string]string)
	stringMap["user_name"] = currentUser.FirstName + " " + currentUser.LastName

	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	render.Template(w, r, "merchant-dashboard.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
	})
}

// Handler for the logout function
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	// Destroy the session
	_ = m.App.Session.Destroy(r.Context())
	m.App.Session.RenewToken(r.Context())
	// Temporary Redirect to the login page
	m.App.Session.Put(r.Context(), "flash", "Logged out succesfully.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// TODO: Fix Authentication for Development mode

// Function to show the administrative verification page
func (m *Repository) ShowAdminVerification(w http.ResponseWriter, r *http.Request) {
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	render.Template(w, r, "merchant-verification.page.tmpl", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// Function to check and validate the verification code:
func (m *Repository) PostShowAdminVerification(w http.ResponseWriter, r *http.Request) {
	// Prevents session attacks
	_ = m.App.Session.RenewToken(r.Context())

	// Get the suer from the session
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	// Add data to the template
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
		return
	}
	// Get the verification code entered by the user
	verificationCode, _ := strconv.Atoi(r.Form.Get("verification_code"))

	// Get the verification code of the user
	dbVRCode, err := m.DB.GetVerificationCode(currentUser)
	if err != nil {
		log.Println("Problem executing the query with error: ", err)
		return
	}

	// Post the form
	form := forms.New(r.PostForm)

	// Perform a check if the verification code is the same
	if verificationCode == dbVRCode {
		// Code is the same
		err := m.DB.IncrementVerification(currentUser)
		if err != nil {
			log.Println("Error in execution of query: ", err)
			return
		}
	} else {
		// Code is not the same
		form.AddVerificationError()
		if !form.Valid() {
			render.Template(w, r, "merchant-verification.page.tmpl", &models.TemplateData{
				Form: form,
				Data: data,
			})
			return
		}
	}

	currentUser.IsVerified++
	m.App.Session.Put(r.Context(), "user_details", currentUser)
	m.App.Session.Put(r.Context(), "flash", "Verification Succesful")

	// Redirect to the address page
	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/verification-address", currentUser.ID), http.StatusSeeOther)
}

// Handler for the show address page
func (m *Repository) ShowAdminAddress(w http.ResponseWriter, r *http.Request) {
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	render.Template(w, r, "merchant-verification-address.page.tmpl", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// PostSHowAddress Handles When the user posts the address
func (m *Repository) PostShowAdminAddress(w http.ResponseWriter, r *http.Request) {
	// Prevents session attacks
	_ = m.App.Session.RenewToken(r.Context())

	// Get the suer from the session
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	// Add data to the template
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	// Parse the form
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	// Make the address
	merchantAddress := models.MerchantAddress{
		City:         r.Form.Get("city"),
		State:        r.Form.Get("state"),
		Country:      r.Form.Get("country"),
		AddressLine1: r.Form.Get("address1"),
		AddressLine2: r.Form.Get("address2"),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Add the address to the database
	err = m.DB.AddMerchantAddress(merchantAddress)
	if err != nil {
		log.Println("ERROR: Error adding merchant address :: error: ", err)
		return
	}

	// Everything is working
	// Increament the verification level by 1
	err = m.DB.IncrementVerification(currentUser)
	if err != nil {
		log.Println("ERROR: Inceamenting Merchant Verification", err)
		return
	}

	// Redirect to a new page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Function to show the documents verification page
func (m *Repository) ShowDocumentsVerification(w http.ResponseWriter, r *http.Request) {
	log.Println("Reached Here")
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	render.Template(w, r, "merchant-verification-documents.page.tmpl", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

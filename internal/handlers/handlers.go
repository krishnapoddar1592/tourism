package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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
	} else if user.IsVerified == 2 {
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

	// Send the verification code through email
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
	// Getting the current User from the session: for the main merchant layout
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
	data["merchant_address"] = models.MerchantAddress{
		City:         "",
		State:        "",
		Country:      "",
		AddressLine1: "",
		AddressLine2: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	render.Template(w, r, "merchant-verification-address.page.tmpl", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// PostSHowAddress Handles When the user posts the address
func (m *Repository) PostShowAdminAddress(w http.ResponseWriter, r *http.Request) {
	// Prevents session attacks
	_ = m.App.Session.RenewToken(r.Context())

	// Get the user from the session
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
		UserID:       currentUser.ID,
	}

	form := forms.New(r.PostForm)
	form.Required("city", "state", "country", "address1", "address2")

	if !form.Valid() {
		data["merchant_address"] = merchantAddress
		render.Template(w, r, "merchant-verification-address.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
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
	// Increment count of the user
	currentUser.IsVerified++
	m.App.Session.Put(r.Context(), "user_details", currentUser)

	// Redirect to a new page
	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/verification-documents", currentUser.ID), http.StatusSeeOther)
}

// Function to show the documents verification page
func (m *Repository) ShowDocumentsVerification(w http.ResponseWriter, r *http.Request) {
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	render.Template(w, r, "merchant-verification-documents.page.tmpl", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) PostShowDocumentsVerification(w http.ResponseWriter, r *http.Request) {
	// Prevents session attacks
	_ = m.App.Session.RenewToken(r.Context())

	// Get the suer from the session
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	// Add data to the template
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	// Dealing with the image first
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Println("Error getting the file", err)
		m.App.Session.Put(r.Context(), "error", "No file was uploaded")
		render.Template(w, r, "merchant-verification-documents.page.tmpl", &models.TemplateData{
			Data: data,
			Form: forms.New(nil),
		})
		return
	}
	defer file.Close()

	if !forms.IsValidFileSize(handler) {
		m.App.Session.Put(r.Context(), "error", "File Size should not be greater than 300 KB")
		render.Template(w, r, "merchant-verification-documents.page.tmpl", &models.TemplateData{
			Data: data,
			Form: forms.New(nil),
		})
		return
	}

	// The final Image to be uploaded to the database
	imageData, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Error loading the image file into bytes")
		return
	}

	// Get the DocumentID as well
	err = r.ParseForm()
	if err != nil {
		log.Println("Error parsing form")
	}
	merchantDocument := models.MerchantDocument{
		DocumentID:   r.Form.Get("documentID"),
		DocumentLink: "testlink",
		ImageFile:    imageData,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		UserID:       currentUser.ID,
	}
	form := forms.New(r.PostForm)
	form.Required("documentID")

	if !form.Valid() {
		render.Template(w, r, "merchant-verification-documents.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Insert merchant document into the database
	documentID, err := m.DB.AddMerchantDocuments(merchantDocument)
	if err != nil {
		log.Println("Error adding merchant documents: ", err)
		return
	}

	// Increament the verification level of the merchant
	err = m.DB.IncrementVerification(currentUser)
	if err != nil {
		log.Println("ERROR: Inceamenting Merchant Verification", err)
		return
	}

	// Add a New Merchant to the Merchants Table
	// 1. Get the Address ID from the User ID
	userAddressID, err := m.DB.GetAddressIDFromUser(currentUser.ID)
	if err != nil {
		log.Println("Error getting address ID: ", err)
		return
	}

	newMerchant := models.MerchantData{
		UserID:     currentUser.ID,
		AddressID:  userAddressID,
		DocumentID: documentID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	// Add the New merchant to the table
	err = m.DB.AddMerchant(newMerchant)
	if err != nil {
		log.Println("Error adding the merchant: ", err)
		return
	}

	mailMessage := fmt.Sprintf(`
		<h2><strong>Email Verification </strong></h2> <br>
		Dear %s, <br>
		<h4>Congratulations! </h4> <br>
		Your Account has succesfully been verified as a Merchant of TourNepal <br>
		<br><br><br>
		Yours Sincerely, <br>
		TourNepal Inc
	`, currentUser.FirstName)

	msg := models.ConfirmationMailData{
		To:      currentUser.Email,
		From:    "info@tournepal.com",
		Subject: "Succesful Account Verification",
		Content: mailMessage,
	}

	// Send the email to the user:
	m.App.MailChan <- msg

	// Destroy the session
	_ = m.App.Session.Destroy(r.Context())
	m.App.Session.RenewToken(r.Context())

	// Redirecting to the merchant dashboard
	m.App.Session.Put(r.Context(), "flash", "Please login again")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Shows an item merchant page
func (m *Repository) AdminAddMerchantItems(w http.ResponseWriter, r *http.Request) {
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	stringMap := make(map[string]string)
	stringMap["user_name"] = currentUser.FirstName + " " + currentUser.LastName

	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	// Get the merchant ID
	merchantID, err := m.DB.GetMerchantIDFromUserID(currentUser.ID)
	if err != nil {
		log.Println("Error getting merchant ID")
		return
	}

	// Add all the portfolio information to the template data

	// Get All the Buses
	buses, err := m.DB.GetAllBus(merchantID)
	if err != nil {
		log.Println("Error getting all the bus data", err)
	}
	data["bus"] = buses
	data["has_bus"] = len(buses)

	// Get all the Hotel Rooms
	rooms, err := m.DB.GetAllHotelRooms(merchantID)
	if err != nil {
		log.Println("Error getting all the room data", err)
	}
	data["hotel_room"] = rooms
	data["has_hotel_room"] = len(rooms)

	render.Template(w, r, "add-merchant-item.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
	})
}

// This funciton shows the add bus page
func (m *Repository) AdminAddBus(w http.ResponseWriter, r *http.Request) {

	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	stringMap := make(map[string]string)
	stringMap["user_name"] = currentUser.FirstName + " " + currentUser.LastName

	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	busDetails := models.AddBusData{
		BusName:     "",
		BusModel:    "",
		BusAddress:  "",
		BusStart:    "",
		BusEnd:      "",
		BusNumSeats: 0,
		BusNumPlate: "",
		BusPAN:      "",
	}

	data["user_details"] = currentUser
	data["bus_reg"] = busDetails
	render.Template(w, r, "add-bus.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// This function handles the Post functionality of the page
func (m *Repository) PostAdminAddBus(w http.ResponseWriter, r *http.Request) {
	// Prevents session attacks
	_ = m.App.Session.RenewToken(r.Context())

	// Get the user from the session
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	// Add data to the template
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	// make stringmap
	stringMap := make(map[string]string)

	// Form Validattion:
	err := r.ParseForm()
	if err != nil {
		log.Println("ERROR: An unexpected Error occured while parsing the form")
	}

	// 1. Form Validation
	// Validate the form
	form := forms.New(r.PostForm)

	// Make the bus details model
	merchantID, err := m.DB.GetMerchantIDFromUserID(currentUser.ID)
	if err != nil {
		log.Println("Error getting merchantID: ", err)
		return
	}

	numSeats := form.ConvertToInt("bus_seats")
	price := form.ConvertToInt("price")

	busDetails := models.AddBusData{
		MerchantID:  merchantID,
		BusName:     r.Form.Get("bus_name"),
		BusModel:    r.Form.Get("bus_model"),
		BusAddress:  r.Form.Get("office_address"),
		BusStart:    r.Form.Get("bus_start"),
		BusEnd:      r.Form.Get("bus_end"),
		BusNumSeats: numSeats,
		BusNumPlate: r.Form.Get("bus_no_plate"),
		BusPAN:      r.Form.Get("bus_pan"),
		Price:       price,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	form.Required("bus_name", "bus_model", "office_address", "bus_start", "bus_end", "bus_seats", "bus_no_plate", "bus_pan", "price")
	form.HasUserAccepted("agreed")

	if !form.Valid() {
		data["bus_reg"] = busDetails
		render.Template(w, r, "add-bus.page.tmpl", &models.TemplateData{
			StringMap: stringMap,
			Data:      data,
			Form:      form,
		})
		return
	}

	// 2. Add the Bus Details to the database
	err = m.DB.AddBusToDatabase(busDetails)
	if err != nil {
		log.Println("Error adding bus details to the bus table")
		return
	}

	// 4. Redirect the User
	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/merchant-add-items", currentUser.ID), http.StatusSeeOther)
}

// This function shows the Records of an individual bus
func (m *Repository) AdminShowBus(w http.ResponseWriter, r *http.Request) {
	// Set up stringmap
	stringMap := make(map[string]string)

	// Add User Details to the session
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	// Add the Bus details in the session
	explodedURL := strings.Split(r.RequestURI, "/")
	busID, _ := strconv.Atoi(explodedURL[4])
	bus, err := m.DB.GetBusByID(busID)
	if err != nil {
		log.Println("Error retrieving bus:", err)
		return
	}
	data["bus_reg"] = bus

	render.Template(w, r, "show-bus.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// This Function Updates the Details of the Merchant Bus
func (m *Repository) PostAdminUpdateBus(w http.ResponseWriter, r *http.Request) {
	// Get current user
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	// Set up stringmap
	stringMap := make(map[string]string)

	// Parse the form
	err := r.ParseForm()
	if err != nil {
		log.Println("Error Parsing the form")
		return
	}

	// Get the Bus ID
	explodedURL := strings.Split(r.RequestURI, "/")
	log.Println(explodedURL)
	busID, _ := strconv.Atoi(explodedURL[4])
	merchantID, _ := strconv.Atoi(explodedURL[2])

	// Get previous bus
	prevBus, err := m.DB.GetBusByID(busID)
	if err != nil {
		log.Println("Couldnt get bus by id : ", err)
		return
	}
	// Post The Form
	form := forms.New(r.PostForm)
	numSeats := form.ConvertToInt("bus_seats")
	price := form.ConvertToInt("price")
	// Form Validation

	// Get the bus
	bus := models.AddBusData{
		BusID:       busID,
		MerchantID:  merchantID,
		BusName:     r.Form.Get("bus_name"),
		BusModel:    r.Form.Get("bus_model"),
		BusAddress:  r.Form.Get("office_address"),
		BusStart:    r.Form.Get("bus_start"),
		BusEnd:      r.Form.Get("bus_end"),
		BusNumSeats: numSeats,
		BusNumPlate: r.Form.Get("bus_no_plate"),
		BusPAN:      r.Form.Get("bus_pan"),
		Price:       price,
		CreatedAt:   prevBus.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	form.Required("bus_name", "bus_model", "office_address", "bus_start", "bus_end", "bus_seats", "bus_no_plate", "bus_pan")
	data := make(map[string]interface{})

	// Add User details to the template
	data["user_details"] = currentUser

	if !form.Valid() {
		log.Println(form.Errors)
		data["bus_reg"] = bus
		render.Template(w, r, "show-bus.page.tmpl", &models.TemplateData{
			StringMap: stringMap,
			Data:      data,
			Form:      form,
		})
		return
	}

	// Update the bus information
	err = m.DB.UpdateBusInfo(busID, bus)
	if err != nil {
		log.Println("Error updating bus information: ", err)
		return
	}

	// Redirect user
	log.Println("Reached Here")
	m.App.Session.Put(r.Context(), "flash", "Changes saved Succesfully!")
	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/merchant-add-items", currentUser.ID), http.StatusSeeOther)
}

// Function to delete the bus
func (m *Repository) PostAdminDeleteBus(w http.ResponseWriter, r *http.Request) {
	// Get current user
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	// Get the bus ID to be deleted
	explodedURL := strings.Split(r.RequestURI, "/")
	busID, _ := strconv.Atoi(explodedURL[5])

	// delete the bus
	err := m.DB.DeleteBusByID(busID)
	if err != nil {
		log.Println("Error in deletion: ", err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Deleted Succesfully")
	// Redirect User
	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/merchant-add-items", currentUser.ID), http.StatusSeeOther)
}

// Function to make the reservation
func (m *Repository) MakeBusReservation(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "make-bus-reservation.page.tmpl", &models.TemplateData{})
}

// TODO: Make this funciton right :: Make this page right
// Function to Post the Bus Reservation to the database
func (m *Repository) PostMakeBusReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("Error parsing the form")
		return
	}

	// Parsing the dates
	resDate, _ := time.Parse("2006-01-02", r.Form.Get("res_date"))
	busID, _ := strconv.Atoi(r.Form.Get("bus_id"))
	numPeople, _ := strconv.Atoi(r.Form.Get("num_people"))

	// Making the data to add to the databse
	busRes := models.BusReservationData{
		BusID:           busID,
		FirstName:       r.Form.Get("first_name"),
		LastName:        r.Form.Get("last_name"),
		ReservationDate: resDate,
		NumPassengers:   numPeople,
		From:            r.Form.Get("from"),
		Stop:            r.Form.Get("to"),
		PhoneNumber:     r.Form.Get("phone"),
		Email:           r.Form.Get("email"),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Submitting to the database
	err = m.DB.MakeBusReservation(busRes)
	if err != nil {
		log.Println("Error adding the reservation to the database: ", err)
		return
	}

	// Redirect to the same page for now
	http.Redirect(w, r, "/make-bus-reservation", http.StatusSeeOther)
}

// Function to show all Bus Reservations
func (m *Repository) ShowAllReservations(w http.ResponseWriter, r *http.Request) {
	// Getting the current User from the session: for the main merchant layout
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	stringMap := make(map[string]string)
	stringMap["user_name"] = currentUser.FirstName + " " + currentUser.LastName

	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	// Get all the reservations from the database
	busRes, err := m.DB.GetAllBusReservations(true)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Add the reservations into the data variable
	data["reservations"] = busRes
	stringMap["is_processed"] = "no"

	render.Template(w, r, "merchant-show-reservations.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
	})
}

// Function to handle when user clicks on a reservation in the table
func (m *Repository) ShowOneReservation(w http.ResponseWriter, r *http.Request) {
	// Getting the current User from the session: for the main merchant layout
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	stringMap := make(map[string]string)
	stringMap["user_name"] = currentUser.FirstName + " " + currentUser.LastName

	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	// Get the Reservation ID
	explodedURL := strings.Split(r.RequestURI, "/")
	id, _ := strconv.Atoi(explodedURL[4])

	// Get Reservation information from ID
	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		log.Println("Error fetching the reservation from the database")
		return
	}
	data["one_res"] = res

	// Send the reservation in the new template and render it
	render.Template(w, r, "merchant-show-busReservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// Function to handle when the bus reservation has been processed:
func (m *Repository) ProcessBusReservation(w http.ResponseWriter, r *http.Request) {
	// Get the reservation ID
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	explodedURL := strings.Split(r.RequestURI, "/")
	id, _ := strconv.Atoi(explodedURL[4])

	err := m.DB.ProcessReservation("bus_reservations", id)
	if err != nil {
		log.Println("There was an error processing the reservation ")
		return
	}

	// TODO: Send a email regarding the processing of the booking

	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/merchant-show-reservations", currentUser.ID), http.StatusSeeOther)
}

// Function to handle when the bus reservation has been deleted:
func (m *Repository) DeleteBusReservation(w http.ResponseWriter, r *http.Request) {
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	explodedURL := strings.Split(r.RequestURI, "/")
	id, _ := strconv.Atoi(explodedURL[4])

	err := m.DB.DeleteBusReservation(id)
	if err != nil {
		log.Println("Error deleting the reservation: ", err)
		return
	}

	// Redirect the user
	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/merchant-show-reservations", currentUser.ID), http.StatusSeeOther)
}

// Function to show all the processed reservations
func (m *Repository) ShowReservationsProcessed(w http.ResponseWriter, r *http.Request) {
	// Getting the current User from the session: for the main merchant layout
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	stringMap := make(map[string]string)
	stringMap["user_name"] = currentUser.FirstName + " " + currentUser.LastName

	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	// Get all the reservations from the database
	busRes, err := m.DB.GetAllBusReservations(false)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Add the reservations into the data variable
	data["reservations"] = busRes
	stringMap["is_processed"] = "yes"

	render.Template(w, r, "merchant-show-reservations.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
	})
}

// Function to handle the posting of an editetd Booking
func (m *Repository) PostShowOneReservation(w http.ResponseWriter, r *http.Request) {
	// Get current user
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	busResUpdate := models.BusReservationData{
		From:        r.Form.Get("from"),
		Stop:        r.Form.Get("stop"),
		PhoneNumber: r.Form.Get("phone"),
		Email:       r.Form.Get("email"),
	}

	// Get the Reservation ID
	explodedURL := strings.Split(r.RequestURI, "/")
	id, _ := strconv.Atoi(explodedURL[4])

	// Update the bus location, email and phone
	err := m.DB.UpdateBusReservation(busResUpdate, id)
	if err != nil {
		log.Println("Error updating the reservation")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/merchant-show-reservations", currentUser.ID), http.StatusSeeOther)
}

// Make a Reservation Calender and display it
func (m *Repository) ShowReservationCalender(w http.ResponseWriter, r *http.Request) {

	// Get the current user
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))

		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	data["user_details"] = currentUser

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)

	stringMap["user_name"] = currentUser.FirstName + " " + currentUser.LastName

	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// get first and last day of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	render.Template(w, r, "merchant-reservation-calender-bus.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})

}

// Function for the merchant to show :: add a hotel
func (m *Repository) AdminAddHotel(w http.ResponseWriter, r *http.Request) {
	// Getting the current User from the session: for the main merchant layout
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	stringMap := make(map[string]string)
	stringMap["user_name"] = currentUser.FirstName + " " + currentUser.LastName

	// Passing the Current User Details to the template data:
	data := make(map[string]interface{})
	data["user_details"] = currentUser
	// Add Empty Hotel Reservation to the template
	data["hotel_reg"] = models.HotelRoom{}

	// Add the reservations into the data variable
	render.Template(w, r, "merchant-add-hotel.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// Post Function for the merchant to post :: add a hotel
func (m *Repository) PostAdminAddHotel(w http.ResponseWriter, r *http.Request) {
	log.Println("This function was called")

	// Prevents session attacks
	_ = m.App.Session.RenewToken(r.Context())

	// Get the user from the session
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	// Add data to the template
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	// make stringmap
	stringMap := make(map[string]string)

	// Server side Form Validation
	err := r.ParseForm()
	if err != nil {
		log.Println("ERROR: An unexpected Error occured while parsing the form")
	}

	// 1. Form Validation
	// Validate the form
	form := forms.New(r.PostForm)

	merchantID, err := m.DB.GetMerchantIDFromUserID(currentUser.ID)
	if err != nil {
		log.Println("Error getting merchantID: ", err)
		return
	}

	numRooms := form.ConvertToInt("no_rooms")
	price := form.ConvertToInt("price")

	// Make the Hotel Reservation Structure
	hotelRoomDetails := models.HotelRoom{
		MerchantID:           merchantID,
		HotelName:            r.Form.Get("hotel_name"),
		HotelRoomName:        r.Form.Get("hotel_room_name"),
		HotelAddress:         r.Form.Get("office_address"),
		HotelType:            r.Form.Get("hotel_type"),
		HotelPAN:             r.Form.Get("hotel_pan"),
		HotelNumRooms:        numRooms,
		HotelPhone1:          r.Form.Get("hotel_phone_1"),
		HotelPhone2:          r.Form.Get("hotel_phone_2"),
		HotelRoomDescription: r.Form.Get("hotel_desc"),
		Price:                price,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// User side form validation
	form.Required("hotel_name", "hotel_room_name", "office_address", "hotel_type", "hotel_pan", "no_rooms", "hotel_phone_1", "hotel_phone_2", "hotel_desc", "price")
	form.Required("agreed")

	if !form.Valid() {
		data["hotel_reg"] = hotelRoomDetails
		render.Template(w, r, "merchant-add-hotel.page.tmpl", &models.TemplateData{
			StringMap: stringMap,
			Data:      data,
			Form:      form,
		})
		return
	}

	// 2. Add the Data To the Database
	err = m.DB.AddNewHotelRoom(hotelRoomDetails)
	if err != nil {
		log.Println("Error inserting the hotel into the database", err)
		return
	}

	// 4. Redirect the user
	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/merchant-add-items", currentUser.ID), http.StatusSeeOther)
}

// TODO: Increase the field size of the Hotel Description

// Show a single Hotel Room Detail
func (m *Repository) AdminShowOneHotel(w http.ResponseWriter, r *http.Request) {
	// Set up stringmap
	stringMap := make(map[string]string)

	// Add User Details to the session
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)
	data := make(map[string]interface{})
	data["user_details"] = currentUser

	// Add the Bus details in the session
	explodedURL := strings.Split(r.RequestURI, "/")
	roomID, _ := strconv.Atoi(explodedURL[4])
	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		log.Println("Error retrieving bus:", err)
		return
	}
	data["hotel_reg"] = room

	render.Template(w, r, "merchant-show-one-hotel.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// Post function to show the hotel
func (m *Repository) PostAdminShowOneHotel(w http.ResponseWriter, r *http.Request) {
	// Get current user
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	// Set up stringmap
	stringMap := make(map[string]string)

	// Parse the form
	err := r.ParseForm()
	if err != nil {
		log.Println("Error Parsing the form")
		return
	}

	// Get the Bus ID
	explodedURL := strings.Split(r.RequestURI, "/")
	log.Println(explodedURL)
	roomID, _ := strconv.Atoi(explodedURL[4])
	merchantID, _ := strconv.Atoi(explodedURL[2])

	// Get previous Room By ID
	prevRoom, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		log.Println("Could not get room by ID: ", err)
		return
	}

	// Post The Form
	form := forms.New(r.PostForm)
	numRooms := form.ConvertToInt("no_rooms")
	price := form.ConvertToInt("price")

	// Form Validation
	hotelRoomDetails := models.HotelRoom{
		MerchantID:           merchantID,
		HotelName:            r.Form.Get("hotel_name"),
		HotelRoomName:        r.Form.Get("hotel_room_name"),
		HotelAddress:         r.Form.Get("office_address"),
		HotelType:            r.Form.Get("hotel_type"),
		HotelPAN:             r.Form.Get("hotel_pan"),
		HotelNumRooms:        numRooms,
		HotelPhone1:          r.Form.Get("hotel_phone_1"),
		HotelPhone2:          r.Form.Get("hotel_phone_2"),
		HotelRoomDescription: r.Form.Get("hotel_desc"),
		Price:                price,
		CreatedAt:            prevRoom.CreatedAt,
		UpdatedAt:            time.Now(),
	}

	// User side form validation
	form.Required("hotel_name", "hotel_room_name", "office_address", "hotel_type", "hotel_pan", "no_rooms", "hotel_phone_1", "hotel_phone_2", "hotel_desc", "price")
	form.Required("agreed")

	data := make(map[string]interface{})
	data["user_details"] = currentUser

	if !form.Valid() {
		log.Println(form.Errors)
		data["hotel_reg"] = hotelRoomDetails

		render.Template(w, r, "merchant-show-one-hotel.page.tmpl", &models.TemplateData{
			StringMap: stringMap,
			Data:      data,
			Form:      form,
		})
	}

	// Update the Bus Information
	err = m.DB.UpdateRoom(hotelRoomDetails, roomID)
	if err != nil {
		log.Println("Error updating room information ", err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Changes saved Succesfully!")
	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/merchant-add-items", currentUser.ID), http.StatusSeeOther)
}

// Function to Delete the Bus
func (m *Repository) DeleteBus(w http.ResponseWriter, r *http.Request) {
	// Get current user
	currentUser := m.App.Session.Get(r.Context(), "user_details").(models.User)

	// Get the bus ID to be deleted
	explodedURL := strings.Split(r.RequestURI, "/")
	roomID, _ := strconv.Atoi(explodedURL[5])

	// Delete the Bus
	err := m.DB.DeleteRoomByID(roomID)
	if err != nil {
		log.Println("Erro rin deletion: ", err)
		return
	}
	m.App.Session.Put(r.Context(), "flash", "Deleted Succesfully")
	// Redirect User
	http.Redirect(w, r, fmt.Sprintf("/merchant/%d/merchant-add-items", currentUser.ID), http.StatusSeeOther)
}

// Function to display the Make Reservation Page for the Hotel Reservations
func (m *Repository) ShowMakeHotelReservation(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "make-hotel-reservation.page.tmpl", &models.TemplateData{})
}

// Function to post the reservation to the database
func (m *Repository) PostShowMakeHotelReservation(w http.ResponseWriter, r *http.Request) {
	log.Println("This funciton is called")
}

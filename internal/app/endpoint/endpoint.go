package endpoint

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/alekslesik/file-cloud/internal/pkg/model"
	"github.com/alekslesik/file-cloud/internal/pkg/templates"
	"github.com/alekslesik/file-cloud/pkg/forms"
	"github.com/alekslesik/file-cloud/pkg/models"
)

type Helpers interface {
	Render(http.ResponseWriter, *http.Request, string, *templates.TemplateData)
	AddDefaultData(*templates.TemplateData, *http.Request) *templates.TemplateData
	AuthenticatedUser(*http.Request) *models.User
}

type ClientServerError interface {
	ClientError(http.ResponseWriter, int, error)
	ServerError(http.ResponseWriter, error)
}

type Endpoint struct {
	h Helpers
	er ClientServerError
	m model.Model
}

func New(h Helpers, er ClientServerError, m model.Model) Endpoint {
	return Endpoint{h: h, er : er, m: m}
}

func (e *Endpoint) home(w http.ResponseWriter, r *http.Request) {
	e.h.Render(w, r, "home.page.html", &templates.TemplateData{})
}

// Login user GET /login.
func (e *Endpoint) loginUserForm(w http.ResponseWriter, r *http.Request) {
	e.h.Render(w, r, "login.page.html", &templates.TemplateData{
		Form: forms.New(nil),
	})
}

// Login user POST /login.
func (e *Endpoint) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		e.er.ClientError(w, http.StatusBadRequest, fmt.Errorf("login user POST /login error"))
		// app.clientError(w, http.StatusBadRequest, fmt.Errorf("login user POST /login error"))
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic
	// message to the form failures map and re-display the login page.
	form := forms.New(r.PostForm)
	id, name, err := e.m.Users.Authenticate(form.Get("email"), form.Get("password"))
	//TODO Add User name to app
	app.UserName = name

	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Email or Password is incorrect")

		e.h.Render(w, r, "login.page.html", &templates.TemplateData{Form: form})
		return
	} else if err != nil {
		e.er.ServerError(w, err)
		return
	}

	// Add the ID of the current user to the session
	app.session.Put(r, "userID", id)

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Sign up user GET /user/signup
func (e *Endpoint) signupUserForm(w http.ResponseWriter, r *http.Request) {
	e.h.Render(w, r, "signup.page.html", &templates.TemplateData{
		Form: forms.New(nil),
	})

}

// Sign up user POST /user/signup
func (e *Endpoint) signupUser(w http.ResponseWriter, r *http.Request) {
	// Parse the form data.
	err := r.ParseForm()
	if err != nil {
		e.er.ClientError(w, http.StatusBadRequest, fmt.Errorf("sign up user POST /user/signup error"))
		return
	}

	// Validate the form contents using the form helper we made earlier.
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 6)

	// If there are any errors, redisplay the signup form.
	if !form.Valid() {
		e.h.Render(w, r, "signup.page.html", &templates.TemplateData{
			Form: form,
		})
		return
	}

	// Try to create a new user record in the database. If the email already exist
	// add an error message to the form and re-display it.
	err = e.m.Users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err == models.ErrDuplicateEmail {
		form.Errors.Add("email", "Address is already in use")
		e.h.Render(w, r, "signup.page.html", &templates.TemplateData{
			Form: form,
		})
		return
	} else if err != nil {
		e.er.ServerError(w, err)
		return
	}

	// Otherwise add a confirmation flash message to the session confirming
	// their signup worked and asking them to log in.
	app.session.Put(r, "flash", "Your signup was successful. Please log in.")

	// GET
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// Logout user GET /user/logout
func (e *Endpoint) logoutUser(w http.ResponseWriter, r *http.Request) {
	// Remove iserID from session.
	app.session.Remove(r, "userID")
	// Add flash to session.
	app.session.Put(r, "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Files page GET /files
func (e *Endpoint) uploadFileForm(w http.ResponseWriter, r *http.Request) {
	// check user authenticate
	if e.h.AuthenticatedUser(r) != nil {
		files, err := app.files.All()
		if err != nil {
			e.er.ServerError(w, err)
		}

		e.h.Render(w, r, "files.page.html", &templates.TemplateData{
			Files: files,
		})
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Files page POST /files
func (e *Endpoint) uploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	r.ParseMultipartForm(32 << 20)

	// Get file from POST
	file, handler, err := r.FormFile("file")
	if err != nil {
		e.er.ClientError(w, http.StatusBadRequest, fmt.Errorf("files page POST /files error"))
	}
	defer file.Close()
	fileType := handler.Header.Get("Content-Type")
	fileName := handler.Filename
	fileSize := handler.Size

	// Try to create a new user record in the database. If the email already exist
	// add an error message to the form and re-display it.
	_, err = e.m.Files.Insert(fileName, fileType, fileSize)
	if err != nil {
		e.er.ServerError(w, err)
		return
	}

	// Create/open file on /upload dir
	f, err := os.OpenFile("../../upload/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		e.er.ServerError(w, err)
	}
	defer f.Close()

	// Write got file to /upload
	io.Copy(f, file)

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/files", http.StatusSeeOther)
}

package endpoint

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/alekslesik/file-cloud/internal/pkg/model"
	"github.com/alekslesik/file-cloud/internal/pkg/session"
	"github.com/alekslesik/file-cloud/internal/pkg/template"
	"github.com/alekslesik/file-cloud/pkg/forms"
	"github.com/alekslesik/file-cloud/pkg/models"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.
const version = "1.0.0"

type Helpers interface {
	Render(http.ResponseWriter, *http.Request, string, *template.TemplateData)
	AddDefaultData(*template.TemplateData, *http.Request) *template.TemplateData
	AuthenticatedUser(*http.Request) *models.User
}

type ClientServerError interface {
	ClientError(http.ResponseWriter, int, error)
	ServerError(http.ResponseWriter, error)
}

type Endpoint struct {
	hp Helpers
	er ClientServerError
	md *model.Model
	ss session.Session
}

func New(hp Helpers, er ClientServerError, md *model.Model, ss session.Session) *Endpoint {
	return &Endpoint{
		hp: hp,
		er: er,
		md: md,
		ss: ss,
	}
}

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (e *Endpoint) HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "version: %s\n", version)
}

// Home GET /
func (e *Endpoint) HomeGet(w http.ResponseWriter, r *http.Request) {
	e.hp.Render(w, r, "home.page.html", &template.TemplateData{})
}

// Login user GET /login.
func (e *Endpoint) UserLoginGet(w http.ResponseWriter, r *http.Request) {
	e.hp.Render(w, r, "login.page.html", &template.TemplateData{
		Form: forms.New(nil),
	})
}

// Login user POST /login.
func (e *Endpoint) UserLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		e.er.ClientError(w, http.StatusBadRequest, fmt.Errorf("login user POST /login error"))
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic
	// message to the form failures map and re-display the login page.
	form := forms.New(r.PostForm)
	id, _, err := e.md.Users.Authenticate(form.Get("email"), form.Get("password"))
	//TODO Add User name to app
	// app.UserName = name

	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Email or Password is incorrect")

		e.hp.Render(w, r, "login.page.html", &template.TemplateData{Form: form})
		return
	} else if err != nil {
		e.er.ServerError(w, err)
		return
	}

	// Add the ID of the current user to the session
	e.ss.Put(r, "userID", id)

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Sign up user GET /user/signup
func (e *Endpoint) UserSignupGet(w http.ResponseWriter, r *http.Request) {
	e.hp.Render(w, r, "signup.page.html", &template.TemplateData{
		Form: forms.New(nil),
	})

}

// Sign up user POST /user/signup
func (e *Endpoint) UserSignupPost(w http.ResponseWriter, r *http.Request) {
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
		e.hp.Render(w, r, "signup.page.html", &template.TemplateData{
			Form: form,
		})
		return
	}

	// Try to create a new user record in the database. If the email already exist
	// add an error message to the form and re-display it.
	err = e.md.Users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err == models.ErrDuplicateEmail {
		form.Errors.Add("email", "Address is already in use")
		e.hp.Render(w, r, "signup.page.html", &template.TemplateData{
			Form: form,
		})
		return
	} else if err != nil {
		e.er.ServerError(w, err)
		return
	}

	// Otherwise add a confirmation flash message to the session confirming
	// their signup worked and asking them to log in.
	e.ss.Put(r, "flash", "Your signup was successful. Please log in.")

	// GET
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// Logout user GET /user/logout
func (e *Endpoint) UserLogoutGet(w http.ResponseWriter, r *http.Request) {
	// Remove userID from session.
	e.ss.Remove(r, "userID")
	// Add flash to session.
	e.ss.Put(r, "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Files page GET /files
func (e *Endpoint) FileUploadGet(w http.ResponseWriter, r *http.Request) {
	// check user authenticate
	if e.hp.AuthenticatedUser(r) != nil {
		files, err := e.md.Files.All()
		if err != nil {
			e.er.ServerError(w, err)
		}

		e.hp.Render(w, r, "files.page.html", &template.TemplateData{
			Files: files,
		})
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Files page POST /files
func (e *Endpoint) FileUploadPost(w http.ResponseWriter, r *http.Request) {
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
	_, err = e.md.Files.Insert(fileName, fileType, fileSize)
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

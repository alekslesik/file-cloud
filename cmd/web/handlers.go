package main

import (
	"io"
	"net/http"
	"os"

	"github.com/alekslesik/file-cloud/pkg/forms"
	"github.com/alekslesik/file-cloud/pkg/models"
)

// Home page GET /
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.html", &templateData{})
}

// Login user GET /login.
func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.html", &templateData{
		Form: forms.New(nil),
	})
}

// Login user POST /login.
func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic
	// message to the form failures map and re-display the login page.
	form := forms.New(r.PostForm)
	id, name, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	// Add User name to app
	app.UserName = name

	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Email or Password is incorrect")

		app.render(w, r, "login.page.html", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// Add the ID of the current user to the session
	app.session.Put(r, "userID", id)

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Sign up user GET /user/signup
func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.html", &templateData{
		Form: forms.New(nil),
	})

}

// Sign up user POST /user/signup
func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	// Parse the form data.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form contents using the form helper we made earlier.
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 6)

	// If there are any errors, redisplay the signup form.
	if !form.Valid() {
		app.render(w, r, "signup.page.html", &templateData{
			Form: form,
		})
		return
	}

	// Try to create a new user record in the database. If the email already exist
	// add an error message to the form and re-display it.
	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err == models.ErrDuplicateEmail {
		form.Errors.Add("email", "Address is already in use")
		app.render(w, r, "signup.page.html", &templateData{
			Form: form,
		})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// Otherwise add a confirmation flash message to the session confirming
	// their signup worked and asking them to log in.
	app.session.Put(r, "flash", "Your signup was successful. Please log in.")

	// GET
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// Logout user GET /user/logout
func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	// Remove iserID from session.
	app.session.Remove(r, "userID")
	// Add flash to session.
	app.session.Put(r, "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Files page GET /files
func (app *application) uploadFileForm(w http.ResponseWriter, r *http.Request) {

	// check user authenticate
	if app.authenticatedUser(r) != nil {
		files, err := app.files.All()
		if err != nil {
			app.serverError(w, err)
		}

		app.render(w, r, "files.page.html", &templateData{
			Files: files,
		})
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Files page POST /files
func (app *application) uploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	r.ParseMultipartForm(32 << 20)

	// Get file from POST
	file, handler, err := r.FormFile("file")
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}
	defer file.Close()
	fileType := handler.Header.Get("Content-Type")
	fileName := handler.Filename
	fileSize := handler.Size

	// Try to create a new user record in the database. If the email already exist
	// add an error message to the form and re-display it.
	_, err = app.files.Insert(fileName, fileType, fileSize)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Create/open file on /upload dir
	f, err := os.OpenFile(app.appPath+"/upload/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		app.serverError(w, err)
	}
	defer f.Close()

	// Write got file to /upload
	io.Copy(f, file)

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/files", http.StatusSeeOther)
}

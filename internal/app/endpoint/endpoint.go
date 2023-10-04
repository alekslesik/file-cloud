package endpoint

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/alekslesik/file-cloud/internal/pkg/mailer"
	"github.com/alekslesik/file-cloud/internal/pkg/model"
	"github.com/alekslesik/file-cloud/internal/pkg/session"
	"github.com/alekslesik/file-cloud/internal/pkg/template"
	"github.com/alekslesik/file-cloud/pkg/forms"
	"github.com/alekslesik/file-cloud/pkg/logging"
	"github.com/alekslesik/file-cloud/pkg/models"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.
const version = "1.0.0"

type Template interface {
	Render(http.ResponseWriter, *http.Request, string, *template.TemplateData)
}

type ClientServerError interface {
	ClientError(http.ResponseWriter, int, error)
	ServerError(http.ResponseWriter, error)
}

type Endpoint struct {
	tmpl Template
	log  *logging.Logger
	er   ClientServerError
	mdl  *model.Model
	ses  session.Session
	mlr  *mailer.Mailer
}

func New(tmpl Template, log *logging.Logger, er ClientServerError, mdl *model.Model, ses session.Session, mlr *mailer.Mailer) *Endpoint {
	return &Endpoint{
		tmpl: tmpl,
		log:  log,
		er:   er,
		mdl:  mdl,
		ses:  ses,
		mlr:  mlr,
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
	flash := e.ses.PopString(r, "flash")
	userName := e.ses.GetString(r, template.UserName)
	e.tmpl.Render(w, r, "home.page.html", &template.TemplateData{
		UserName: userName,
		Flash:    flash,
	})
}

// Login user GET /login.
func (e *Endpoint) UserLoginGet(w http.ResponseWriter, r *http.Request) {
	flash := e.ses.PopString(r, "flash")
	e.tmpl.Render(w, r, "login.page.html", &template.TemplateData{
		Flash: flash,
		Form:  forms.New(nil),
	})
}

// Login user POST /login.
func (e *Endpoint) UserLoginPost(w http.ResponseWriter, r *http.Request) {
	const op = "endpoint.UserLoginPost()"

	err := r.ParseForm()
	if err != nil {
		e.er.ClientError(w, http.StatusBadRequest, fmt.Errorf("login user POST /login error"))
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic
	// message to the form failures map and re-display the login page.
	form := forms.New(r.PostForm)

	email := form.Get("email")
	password := form.Get("password")
	id, userName, err := e.mdl.Users.Authenticate(email, password)

	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Email or Password is incorrect")
		td := &template.TemplateData{}
		td.Form = form
		e.tmpl.Render(w, r, "login.page.html", td)
		return
	} else if err != nil {
		e.er.ServerError(w, err)
		return
	}

	// Add the ID of the current user to the session
	e.ses.Put(r, template.UserID, id)
	e.ses.Put(r, template.UserName, userName)

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Sign up user GET /user/signup
func (e *Endpoint) UserSignupGet(w http.ResponseWriter, r *http.Request) {
	e.tmpl.Render(w, r, "signup.page.html", &template.TemplateData{
		Form: forms.New(nil),
	})
}

// Sign up user POST /user/signup
func (e *Endpoint) UserSignupPost(w http.ResponseWriter, r *http.Request) {
	const op = "endpoint.UserSignupPost()"

	// Parse the form data.
	err := r.ParseForm()
	if err != nil {
		e.log.Err(err).Msgf("%s > parse form", op)
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
		errMsg := form.Errors.WholeErrorMessage("<br>")
		form.Errors.Add("generic", errMsg)
		td := &template.TemplateData{}
		td.Form = form
		e.tmpl.Render(w, r, "signup.page.html", td)
		return
	}

	// Try to create a new user record in the database. If the email already exist
	// add an error message to the form and re-display it.
	err = e.mdl.Users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err == models.ErrDuplicateEmail {
		e.log.Err(err).Msgf("%s > duplicate email", op)
		form.Errors.Add("email", "Address is already in use")
		e.tmpl.Render(w, r, "signup.page.html", &template.TemplateData{
			Form: form,
		})
		return
	} else if err != nil {
		e.log.Err(err).Msgf("%s > user insert", op)
		e.er.ServerError(w, err)
		return
	}

	// Otherwise add a confirmation flash message to the session confirming
	// their signup worked and asking them to log in.
	e.ses.Put(r, "flash", "Your signup was successful. Please log in.")

	// Call the Send() method on our Mailer, passing in the user's email address,
	// name of the template file, and the User struct containing the new user's data.
	email := form.Get("email")
	name := form.Get("name")
	err = e.mlr.Send(email, "user_welcome.html", struct{ Name string }{Name: name})
	if err != nil {
		e.log.Err(err).Msgf("%s > mail send error", op)
		e.er.ServerError(w, err)
		return
	}

	// GET
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// Logout user GET /user/logout
func (e *Endpoint) UserLogoutGet(w http.ResponseWriter, r *http.Request) {
	// Remove userID from session.
	e.ses.Remove(r, template.UserID)
	e.ses.Remove(r, template.UserName)
	// Add flash to session.
	e.ses.Put(r, "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Files page GET /files
func (e *Endpoint) FileUploadGet(w http.ResponseWriter, r *http.Request) {
	const op = "endpoint.FileUploadGet()"

	userId := e.ses.GetInt(r, template.UserID)

	// check user authenticate
	if template.AuthenticatedUser(r) != nil {
		files, err := e.mdl.Files.All(userId)
		if err != nil {
			e.log.Err(err).Msgf("%s > get all files from DB", op)
			e.er.ServerError(w, err)
			return
		}

		userName := e.ses.GetString(r, template.UserName)
		e.tmpl.Render(w, r, "files.page.html", &template.TemplateData{
			UserName: userName,
			Files:    files,
		})
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Files page POST /files
func (e *Endpoint) FileUploadPost(w http.ResponseWriter, r *http.Request) {
	const op = "endpoint.FileUploadPost()"

	// Parse request body
	r.ParseMultipartForm(32 << 20)

	// Get file from POST
	file, fHeader, err := r.FormFile("file")
	if err != nil {
		e.log.Err(err).Msgf("%s > get file from post", op)
		e.er.ClientError(w, http.StatusBadRequest, fmt.Errorf("files page POST /files error"))
	}
	defer file.Close()
	fileType := fHeader.Header.Get("Content-Type")
	fileName := fHeader.Filename
	fileSize := fHeader.Size
	fileURL := "/download/" + fHeader.Filename

	userId := e.ses.GetInt(r, template.UserID)

	// Try to create a new user record in the database. If the email already exist
	// add an error message to the form and re-display it.
	_, err = e.mdl.Files.Insert(fileName, fileType, fileSize, fileURL, userId)
	if err != nil {
		e.log.Err(err).Msgf("%s > insert file to DB", op)
		e.er.ServerError(w, err)
		return
	}

	newFile := "./website/upload/" + fHeader.Filename

	// Create/open file from /upload dir
	f, err := os.OpenFile(newFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		e.log.Err(err).Msgf("%s > open file", op)
		e.er.ServerError(w, err)
	}
	defer f.Close()

	// Write got file to /upload
	io.Copy(f, file)

	//

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/files", http.StatusSeeOther)
}

// Download page GET /download/:filename
func (e *Endpoint) FileDownloadGet(w http.ResponseWriter, r *http.Request) {
	const op = "endpoint.FileDownloadGet()"

	// Get file name from URL parameter
	filename := r.URL.Query().Get(":filename")

	// Full filepath on server
	filePath := "./website/upload/" + filename

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		e.log.Err(err).Msgf("%s > open file", op)
		e.er.ServerError(w, err)
	}
	defer file.Close()

	// Set header for downloading file
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	// Send file
	io.Copy(w, file)
}

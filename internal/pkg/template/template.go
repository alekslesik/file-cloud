package template

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"path/filepath"
	"time"

	"github.com/alekslesik/file-cloud/pkg/forms"
	"github.com/alekslesik/file-cloud/pkg/logging"
	"github.com/alekslesik/file-cloud/pkg/models"
	"github.com/alekslesik/file-cloud/internal/pkg/helpers"
	"github.com/justinas/nosurf"
)

type ClientServerError interface {
	ClientError(http.ResponseWriter, int, error)
	ServerError(http.ResponseWriter, error)
}

type Cache map[string]*template.Template

type Template struct {
	tmpl  *TemplateData
	cache Cache
	log   *logging.Logger
	er ClientServerError
}

type TemplateData struct {
	AuthenticatedUser *models.User
	UserName          string
	Flash             string
	CurrentYear       int
	CSRFToken         string
	Form              *forms.Form
	File              *models.File
	Files             []*models.File
}

func New(logger *logging.Logger) *Template {
	return &Template{
		tmpl:  new(TemplateData),
		cache: make(Cache),
		log:   logger,
	}
}

// Return nicely formatted string of time.Time object
func HumanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	// Convert the time to UTC before formatting it.
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// Initialize a template.FuncMap object and store it in a global variable. This
// essentially a string-keyed map which acts as a lookup between the names of o
// custom template functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": HumanDate,
}

// Add template cache of files in dir
func (t *Template) NewCache(dir string) *Template {
	const op = "template.NewCache()"

	cache, err := t.newCache(dir)
	if err != nil {
		t.log.Err(err).Msgf("%s: open db", op)
	}

	t.cache = cache
	return t
}

func (t *Template) newCache(dir string) (Cache, error) {
	const op = "template.newCache()"

	// init new map keeping cache
	cache := map[string]*template.Template{}

	// use func Glob to get all filepathes slice with '.page.html' ext
	entries, err := filepath.Glob(filepath.Join(dir, "*.page.html"))
	if err != nil {
		t.log.Err(err).Msgf("%s: glob *.page.html in dir %v", op, dir)
		return nil, err
	}

	for _, e := range entries {
		// get filename from filepath
		name := filepath.Base(e)

		// The template.FuncMap must be registered with the template set before
		// call the ParseFiles() method. This means we have to use template.New
		// create an empty template set, use the Funcs() method t
		ts, err := template.New(name).Funcs(functions).ParseFiles(e)
		if err != nil {
			t.log.Err(err).Msgf("%s: template create", op)
			return nil, err
		}

		// use ParseGlob to add all frame patterns (base.layout.html)
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.html"))
		if err != nil {
			t.log.Err(err).Msgf("%s: glob *.layout.html to template", op)
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.html"))
		if err != nil {
			t.log.Err(err).Msgf("%s: glob *.partial.html to template", op)
			return nil, err
		}

		// add received patterns set to cache, using page name
		// (ext home.page.html) as a key for our map
		cache[name] = ts
	}

	return cache, nil
}

func (t *Template) Render(w http.ResponseWriter, r *http.Request, name string, td *TemplateData) {
	const op = "helpers.Render()"

	// extract pattern depending "name"
	ts, ok := t.cache[name]
	if !ok {
		t.log.Error().Msgf("%s > pattern %s not exist", op, name)
		t.er.ServerError(w, fmt.Errorf("pattern %s not exist", name))
		return
	}

	// initialize a new buffer
	buf := new(bytes.Buffer)

	// write template to the buffer, instead straight to http.ResponseWriter
	err := ts.Execute(buf, AddDefaultData(td, r))
	if err != nil {
		t.log.Error().Msgf("%s > template %v not executed", op, ts)
		t.er.ServerError(w, fmt.Errorf("template %v not executed", ts))
		return
	}

	// write buffer to http.ResponseWriter
	buf.WriteTo(w)
}

// Create an addDefaultData helper. This takes a pointer to a templateData
// struct, adds the current year to the CurrentYear field, and then returns
// the pointer. Again, we're not using the *http.Request parameter at the
// moment, but we will do later in the book.
func AddDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	if td == nil {
		td = &TemplateData{}
	}

	// Add current time.
	td.CurrentYear = time.Now().Year()
	// Add flash message.
	// TODO sort out
	// td.Flash = app.session.PopString(r, "flash")
	// Check if user is authenticate.
	td.AuthenticatedUser = helpers.AuthenticatedUser(r)
	// Add the CSRF token to the templateData struct.
	td.CSRFToken = nosurf.Token(r)
	// Add User Name to template
	// td.UserName = app.UserName

	return td
}
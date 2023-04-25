package main

import (
	"embed"

	"html/template"
	"path"

	// "path"
	"path/filepath"
	"time"

	"github.com/alekslesik/file-cloud/pkg/forms"
	"github.com/alekslesik/file-cloud/pkg/models"
	// "github.com/rs/zerolog/log"
)

type templateData struct {
	AuthenticatedUser *models.User
	UserName          string
	Flash             string
	CurrentYear       int
	CSRFToken         string
	Form              *forms.Form
	File              *models.File
	Files             []*models.File
}

// Below we declare a new variable with the type embed.FS (embedded file system) to hold
// our email templates. This has a comment directive in the format `//go:embed <path>`
// IMMEDIATELY ABOVE it, which indicates to Go that we want to store the contents of the
// ./templates directory in the templateFS embedded file system variable.
// ↓↓↓

//go:embed ui/html/*.html
var embedFS embed.FS

// Return nicely formatted string of time.Time object
func humanDate(t time.Time) string {
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
	"humanDate": humanDate,
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// init new map keeping cache
	cache := map[string]*template.Template{}

	// take all file from embed FS
	readDir, err := embedFS.ReadDir("ui/html")
	if err != nil {
		return nil, err
	}

	for _, page := range readDir {
		// get filename from filepath
		name := filepath.Base(page.Name())

		// The template.FuncMap must be registered with the template set before
		// call the ParseFiles() method. This means we have to use template.New
		// create an empty template set, use the Funcs() method t
		// ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		lp := path.Join(dir, "*.layout.html")
		fp := path.Join(dir, name)
		pp := path.Join(dir, "*.partial.html")

		ts, err := template.New(name).Funcs(functions).ParseFS(embedFS, lp, fp, pp)
		if err != nil {
			return nil, err
		}

		// add received patterns set to cache, using page name
		// (ext home.page.html) as a key for our map
		cache[name] = ts
	}

	return cache, nil
}

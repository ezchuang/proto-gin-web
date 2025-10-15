package templates

import (
	"errors"
	"html/template"
	"path/filepath"

	"github.com/gin-contrib/multitemplate"
)

// MustLoad combines the provided template files (layout first) into a single *template.Template.
// It panics if parsing fails, mirroring the behaviour of template.Must.
func MustLoad(files ...string) *template.Template {
	if len(files) == 0 {
		panic("templates: at least one template file is required")
	}
	absFiles := make([]string, len(files))
	for i, f := range files {
		absFiles[i] = filepath.Clean(f)
	}
	return template.Must(template.ParseFiles(absFiles...))
}

// Load parses the given template files and returns the compiled template.
// The caller can handle the returned error to surface template issues gracefully.
func Load(files ...string) (*template.Template, error) {
	if len(files) == 0 {
		return nil, ErrNoTemplateFiles
	}
	absFiles := make([]string, len(files))
	for i, f := range files {
		absFiles[i] = filepath.Clean(f)
	}
	return template.ParseFiles(absFiles...)
}

// ErrNoTemplateFiles indicates that no template files were provided to Load.
var ErrNoTemplateFiles = errors.New("templates: no template files provided")

// MustParseGlob parses all templates matching the provided pattern and panics if parsing fails.
func MustParseGlob(pattern string) *template.Template {
	t, err := template.ParseGlob(filepath.Clean(pattern))
	if err != nil {
		panic(err)
	}
	return t
}

// ParseGlob parses the templates matching pattern and returns them for manual error handling.
func ParseGlob(pattern string) (*template.Template, error) {
	return template.ParseGlob(filepath.Clean(pattern))
}

// LoadTemplates builds a Gin multitemplate renderer using layout & include folders.
func LoadTemplates(templatesDir string, layoutPattern string, includePattern string) multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	layouts, err := filepath.Glob(filepath.Join(templatesDir, layoutPattern))
	if err != nil {
		panic(err)
	}
	includes, err := filepath.Glob(filepath.Join(templatesDir, includePattern))
	if err != nil {
		panic(err)
	}

	for _, include := range includes {
		files := append([]string{}, layouts...)
		files = append(files, include)
		r.AddFromFiles(filepath.Base(include), files...)
	}

	return r
}

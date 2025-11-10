package templates

import (
	"path/filepath"

	"github.com/gin-contrib/multitemplate"
)

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

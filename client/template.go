package main

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type Template struct {
	templates map[string]*template.Template
}

func NewTemplate() (t *Template, err error) {

	t = &Template{
		templates: make(map[string]*template.Template),
	}

	if err := t.parseTemplates(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Template) parseTemplates() error {
	// Get all layout files
	layoutFiles, err := filepath.Glob("public/views/layout/*.html")
	if err != nil {
		return fmt.Errorf("error finding layout files: %w", err)
	}

	// Get all partial files
	partialFiles, err := filepath.Glob("public/views/partials/*.html")
	if err != nil {
		return fmt.Errorf("error finding partial files: %w", err)
	}

	// Get all page files from multiple directories
	pageFiles, err := t.getAllPageFiles()
	if err != nil {
		return fmt.Errorf("error finding page files: %w", err)
	}

	// For each page, create a template that includes layouts and partials
	for _, pageFile := range pageFiles {
		// Get the base name of the page file (without extension)
		pageName := filepath.Base(pageFile)
		templateName := pageName[:len(pageName)-len(filepath.Ext(pageName))]

		// Combine all template files for this page
		var templateFiles []string
		templateFiles = append(templateFiles, layoutFiles...)
		templateFiles = append(templateFiles, partialFiles...)
		templateFiles = append(templateFiles, pageFile)

		funcMap := template.FuncMap{}

		// Parse the combined templates
		tmpl, err := template.New(templateName).Funcs(funcMap).ParseFiles(templateFiles...)
		if err != nil {
			return fmt.Errorf("error parsing template %s: %w", templateName, err)
		}

		t.templates[templateName] = tmpl
	}

	return nil
}

// Helper function to collect all page files
func (t *Template) getAllPageFiles() ([]string, error) {
	var allFiles []string
	
	// Define all the directories where page templates can be found
	pageDirs := []string{
		"public/views/auth/*.html",
		"public/views/public/*.html", 
		"public/views/app/*.html",
		"public/views/page/*.html", // Legacy support
	}
	
	for _, pattern := range pageDirs {
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("error finding files with pattern %s: %w", pattern, err)
		}
		allFiles = append(allFiles, files...)
	}
	
	return allFiles, nil
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, exists := t.templates[name]
	if !exists {
		return fmt.Errorf("template %s not found", name)
	}

	return tmpl.Execute(w, data)
}
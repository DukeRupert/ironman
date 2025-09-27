package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed public/views/*
var templateFS embed.FS

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
	layoutFiles, err := t.getFilesFromEmbedded("public/views/layout/*.html")
	if err != nil {
		return fmt.Errorf("error finding layout files: %w", err)
	}

	// Get all partial files
	partialFiles, err := t.getFilesFromEmbedded("public/views/partials/*.html")
	if err != nil {
		return fmt.Errorf("error finding partial files: %w", err)
	}

	// Get all page files from multiple directories
	pageFiles, err := t.getAllPageFilesFromEmbedded()
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

		// caser for English
		titleCaser := cases.Title(language.English)
		upperCaser := cases.Upper(language.English)
		lowerCaser := cases.Lower(language.AmericanEnglish)
		funcMap := template.FuncMap{
			// Math functions
			"add": func(a, b int) int { return a + b },
			"sub": func(a, b int) int { return a - b },
			"mul": func(a, b float64) float64 { return a * b },
			"div": func(a, b float64) float64 {
				if b == 0 {
					return 0
				}
				return a / b
			},

			// String functions
			"upper": func(s string) string { return upperCaser.String(s) },
			"lower": func(s string) string { return lowerCaser.String(s) },
			"title": func(s string) string { return titleCaser.String(s) },

			// Length functions
			"len": func(v interface{}) int {
				switch val := v.(type) {
				case []interface{}:
					return len(val)
				case string:
					return len(val)
				default:
					return 0
				}
			},
		}

		// Parse the combined templates from embedded filesystem
		tmpl := template.New(templateName).Funcs(funcMap)
		
		// Parse each template file from the embedded filesystem
		for _, file := range templateFiles {
			content, err := templateFS.ReadFile(file)
			if err != nil {
				return fmt.Errorf("error reading embedded template %s: %w", file, err)
			}
			
			// Parse the template content
			_, err = tmpl.Parse(string(content))
			if err != nil {
				return fmt.Errorf("error parsing template %s: %w", file, err)
			}
		}

		t.templates[templateName] = tmpl
	}

	return nil
}

// Helper function to get files matching a pattern from embedded filesystem
func (t *Template) getFilesFromEmbedded(pattern string) ([]string, error) {
	var files []string
	
	// Extract directory from pattern
	dir := filepath.Dir(pattern)
	
	// Walk through the embedded filesystem
	err := fs.WalkDir(templateFS, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if d.IsDir() {
			return nil
		}
		
		// Check if file matches pattern
		matched, err := filepath.Match(filepath.Base(pattern), filepath.Base(path))
		if err != nil {
			return err
		}
		
		if matched {
			files = append(files, path)
		}
		
		return nil
	})
	
	return files, err
}

// Helper function to collect all page files from embedded filesystem
func (t *Template) getAllPageFilesFromEmbedded() ([]string, error) {
	var allFiles []string

	// Define all the directories where page templates can be found
	pageDirs := []string{
		"public/views/pages/auth",
		"public/views/pages/app", 
		"public/views/pages",
	}

	for _, dir := range pageDirs {
		// Check if directory exists in embedded filesystem
		entries, err := fs.ReadDir(templateFS, dir)
		if err != nil {
			// Skip if directory doesn't exist
			continue
		}
		
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			
			// Only include .html files
			if strings.HasSuffix(entry.Name(), ".html") {
				fullPath := filepath.Join(dir, entry.Name())
				allFiles = append(allFiles, fullPath)
			}
		}
	}

	return allFiles, nil
}

func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	tmpl, exists := t.templates[name]
	if !exists {
		return fmt.Errorf("template %s not found", name)
	}

	return tmpl.Execute(w, data)
}

package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/yourusername/event-feedback/internal/database"
	"github.com/yourusername/event-feedback/internal/utils"
	"gorm.io/gorm"
)

var (
	// Templates holds all parsed templates
	Templates map[string]*template.Template
	// DB is the database connection
	DB *gorm.DB
)

// RegisterHandlers registers all HTTP handlers
func RegisterHandlers(mux *http.ServeMux, db *sql.DB) {
	// Use the GORM DB instance directly
	DB = database.DB

	// Parse templates
	parseTemplates()

	// Home page
	mux.HandleFunc("/", HomeHandler)

	// Event related routes
	mux.HandleFunc("/events", ListEventsHandler)
	mux.HandleFunc("/events/new", NewEventHandler)
	mux.HandleFunc("/events/create", CreateEventHandler)
	mux.HandleFunc("/events/view/", ViewEventHandler)

	// Form related routes
	mux.HandleFunc("/forms/new/", NewFormHandler)
	mux.HandleFunc("/forms/create", CreateFormHandler)
	mux.HandleFunc("/forms/edit/", EditFormHandler)
	mux.HandleFunc("/forms/update", UpdateFormHandler)
	mux.HandleFunc("/forms/view/", ViewFormHandler)
	mux.HandleFunc("/forms/submit/", SubmitFormHandler)

	// Submission related routes
	mux.HandleFunc("/submissions/view/", ViewSubmissionHandler)
	mux.HandleFunc("/submissions/continue/", ContinueSubmissionHandler)
}

// parseTemplates parses all HTML templates
func parseTemplates() {
	Templates = make(map[string]*template.Template)

	// Define the layout and template directory
	layoutFile := filepath.Join("internal", "templates", "layout.html")
	templateDir := filepath.Join("internal", "templates")

	// Get template functions
	funcMap := utils.TemplateFuncs()

	// Get all template files
	templateFiles, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		panic(err)
	}

	// Parse each template with the layout
	for _, templateFile := range templateFiles {
		if templateFile == layoutFile {
			continue
		}

		fileName := filepath.Base(templateFile)
		tmpl := template.New(fileName).Funcs(funcMap)
		tmpl = template.Must(tmpl.ParseFiles(layoutFile, templateFile))
		Templates[fileName] = tmpl
	}
}

// RenderTemplate renders a template with the given data
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	template, ok := Templates[tmpl]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	err := template.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

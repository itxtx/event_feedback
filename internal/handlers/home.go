package handlers

import (
	"net/http"

	"github.com/yourusername/event-feedback/internal/database"
	"gorm.io/gorm"
)

// PageData holds common data for all pages
type PageData struct {
	Title      string
	Error      string
	Success    string
	Events     []database.Event
	Event      database.Event
	Form       database.Form
	Forms      []database.Form
	Fields     []database.FormField
	Submission database.Submission
}

// HomeHandler handles the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get recent events
	var events []database.Event
	result := database.DB.Order("created_at desc").Limit(5).Find(&events)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	// Get published forms with their events
	var forms []database.Form
	result = database.DB.
		Where("is_published = ?", true).
		Order("created_at desc").
		Limit(10).
		Preload("Event").
		Find(&forms)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		http.Error(w, "Failed to fetch forms", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:  "Event Feedback System",
		Events: events,
		Forms:  forms,
	}

	RenderTemplate(w, "home.html", data)
}

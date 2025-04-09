package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/event-feedback/internal/database"
	"gorm.io/gorm"
)

// ListEventsHandler handles listing all events
func ListEventsHandler(w http.ResponseWriter, r *http.Request) {
	var events []database.Event
	result := database.DB.Order("date desc").Find(&events)
	if result.Error != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:  "All Events",
		Events: events,
	}

	RenderTemplate(w, "events.html", data)
}

// NewEventHandler displays the form to create a new event
func NewEventHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Create New Event",
	}

	RenderTemplate(w, "new_event.html", data)
}

// CreateEventHandler handles the form submission to create a new event
func CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	name := r.FormValue("name")
	description := r.FormValue("description")
	dateStr := r.FormValue("date")

	// Validate required fields
	if name == "" || dateStr == "" {
		data := PageData{
			Title: "Create New Event",
			Error: "Name and date are required",
		}
		RenderTemplate(w, "new_event.html", data)
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		data := PageData{
			Title: "Create New Event",
			Error: "Invalid date format",
		}
		RenderTemplate(w, "new_event.html", data)
		return
	}

	// Create new event
	event := database.Event{
		Name:        name,
		Description: description,
		Date:        date,
	}

	result := database.DB.Create(&event)
	if result.Error != nil {
		data := PageData{
			Title: "Create New Event",
			Error: "Failed to create event: " + result.Error.Error(),
		}
		RenderTemplate(w, "new_event.html", data)
		return
	}

	// Redirect to view the new event
	http.Redirect(w, r, "/events/view/"+strconv.FormatUint(uint64(event.ID), 10), http.StatusSeeOther)
}

// ViewEventHandler displays an event and its forms
func ViewEventHandler(w http.ResponseWriter, r *http.Request) {
	// Extract event ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/events/view/")
	id, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Get event
	var event database.Event
	result := database.DB.First(&event, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
		}
		return
	}

	// Get forms for this event
	var forms []database.Form
	result = database.DB.Where("event_id = ?", id).Find(&forms)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		http.Error(w, "Failed to fetch forms", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title: event.Name,
		Event: event,
		Forms: forms,
	}

	RenderTemplate(w, "view_event.html", data)
}

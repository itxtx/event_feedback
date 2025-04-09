package handlers

import (
	"net/http"
	"strconv"

	"github.com/yourusername/event-feedback/internal/database"
	"gorm.io/gorm"
)

// ViewSubmissionHandler displays a submitted form
func ViewSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	// Extract submission key from URL
	submissionKey := r.URL.Path[len("/submissions/view/"):]
	if submissionKey == "" {
		http.NotFound(w, r)
		return
	}

	// Get submission
	var submission database.Submission
	result := database.DB.Where("submission_key = ?", submissionKey).First(&submission)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Failed to fetch submission", http.StatusInternalServerError)
		}
		return
	}

	// Get form
	var form database.Form
	result = database.DB.First(&form, submission.FormID)
	if result.Error != nil {
		http.Error(w, "Failed to fetch form", http.StatusInternalServerError)
		return
	}

	// Get event
	var event database.Event
	result = database.DB.First(&event, form.EventID)
	if result.Error != nil {
		http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
		return
	}

	// Get all responses with field information
	type ResponseWithField struct {
		database.SubmissionResponse
		Step      int
		FieldType string
		Label     string
	}

	var responses []ResponseWithField
	result = database.DB.Model(&database.SubmissionResponse{}).
		Select("submission_responses.*, form_fields.step, form_fields.field_type, form_fields.label").
		Joins("LEFT JOIN form_fields ON form_fields.id = submission_responses.field_id").
		Where("submission_responses.submission_id = ?", submission.ID).
		Order("form_fields.step, form_fields.field_order").
		Scan(&responses)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		http.Error(w, "Failed to fetch responses", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:      "Submission: " + form.Title,
		Form:       form,
		Event:      event,
		Submission: submission,
	}

	// Add responses to page data
	type ResponseData struct {
		Label     string
		Value     string
		Step      int
		FieldType string
	}

	responseData := make([]ResponseData, 0, len(responses))
	for _, resp := range responses {
		responseData = append(responseData, ResponseData{
			Label:     resp.Label,
			Value:     resp.Response,
			Step:      resp.Step,
			FieldType: resp.FieldType,
		})
	}

	// Add responses to page data using a template variable
	tmpl := Templates["view_submission.html"]
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	// Execute template with responses
	err := tmpl.ExecuteTemplate(w, "layout", struct {
		PageData
		Responses []ResponseData
	}{
		PageData:  data,
		Responses: responseData,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ContinueSubmissionHandler allows users to continue an in-progress submission
func ContinueSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	// Extract submission key from URL
	submissionKey := r.URL.Path[len("/submissions/continue/"):]
	if submissionKey == "" {
		http.NotFound(w, r)
		return
	}

	// Get submission
	var submission database.Submission
	result := database.DB.Where("submission_key = ?", submissionKey).First(&submission)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Failed to fetch submission", http.StatusInternalServerError)
		}
		return
	}

	// Check if submission is already completed
	if submission.Status == "completed" {
		http.Redirect(w, r, "/submissions/view/"+submissionKey, http.StatusSeeOther)
		return
	}

	// Get form
	var form database.Form
	result = database.DB.First(&form, submission.FormID)
	if result.Error != nil {
		http.Error(w, "Failed to fetch form", http.StatusInternalServerError)
		return
	}

	// Check if form is still published
	if !form.IsPublished {
		http.Error(w, "This form is no longer available", http.StatusForbidden)
		return
	}

	// Get event
	var event database.Event
	result = database.DB.First(&event, form.EventID)
	if result.Error != nil {
		http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
		return
	}

	// Get form fields for the current step
	var fields []database.FormField
	result = database.DB.Where("form_id = ? AND step = ?", form.ID, submission.CurrentStep).
		Order("field_order").Find(&fields)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		http.Error(w, "Failed to fetch form fields", http.StatusInternalServerError)
		return
	}

	// Get previous responses to prefill the form
	for i := range fields {
		var response database.SubmissionResponse
		result = database.DB.Where("submission_id = ? AND field_id = ?", submission.ID, fields[i].ID).
			First(&response)

		if result.Error == nil {
			// Create a custom attribute to hold the previous response
			// This will be used in the template to prefill the form fields
			fields[i].Placeholder = response.Response
		}
	}

	data := PageData{
		Title:      form.Title + " - Continue from Step " + strconv.Itoa(submission.CurrentStep),
		Form:       form,
		Event:      event,
		Fields:     fields,
		Submission: submission,
	}

	RenderTemplate(w, "view_form.html", data)
}

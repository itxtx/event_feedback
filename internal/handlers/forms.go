package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/event-feedback/internal/database"
	"gorm.io/gorm"
)

// NewFormHandler displays the form to create a new feedback form
func NewFormHandler(w http.ResponseWriter, r *http.Request) {
	// Extract event ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/forms/new/")
	eventID, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	
	// Redirect back to the form editor
	http.Redirect(w, r, "/forms/edit/"+formIDStr, http.StatusSeeOther)
}

// ViewFormHandler displays a form for a user to fill out
func ViewFormHandler(w http.ResponseWriter, r *http.Request) {
	// Extract form ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/forms/view/")
	formID, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	
	// Get form
	var form database.Form
	result := database.DB.First(&form, formID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Failed to fetch form", http.StatusInternalServerError)
		}
		return
	}
	
	// Check if form is published
	if !form.IsPublished {
		http.Error(w, "This form is not available", http.StatusForbidden)
		return
	}
	
	// Get event
	var event database.Event
	result = database.DB.First(&event, form.EventID)
	if result.Error != nil {
		http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
		return
	}
	
	// Get form fields for the first step (or all if not multi-step)
	var fields []database.FormField
	query := database.DB.Where("form_id = ?", formID)
	
	if form.IsMultiStep {
		query = query.Where("step = ?", 1)
	}
	
	result = query.Order("field_order").Find(&fields)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		http.Error(w, "Failed to fetch form fields", http.StatusInternalServerError)
		return
	}
	
	// Create a new submission
	submissionKey := generateSubmissionKey()
	submission := database.Submission{
		FormID:        uint(formID),
		SubmissionKey: submissionKey,
		Status:        "in_progress",
		CurrentStep:   1,
	}
	
	result = database.DB.Create(&submission)
	if result.Error != nil {
		http.Error(w, "Failed to create submission: "+result.Error.Error(), http.StatusInternalServerError)
		return
	}
	
	data := PageData{
		Title:      form.Title,
		Form:       form,
		Event:      event,
		Fields:     fields,
		Submission: submission,
	}
	
	RenderTemplate(w, "view_form.html", data)
}

// SubmitFormHandler handles the form submission from users
func SubmitFormHandler(w http.ResponseWriter, r *http.Request) {
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
	
	// Get submission ID and form ID
	submissionIDStr := r.FormValue("submission_id")
	formIDStr := r.FormValue("form_id")
	currentStepStr := r.FormValue("current_step")
	
	// Validate required fields
	if submissionIDStr == "" || formIDStr == "" || currentStepStr == "" {
		http.Error(w, "Submission ID, form ID, and current step are required", http.StatusBadRequest)
		return
	}
	
	// Parse IDs and step
	submissionID, err := strconv.ParseUint(submissionIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}
	
	formID, err := strconv.ParseUint(formIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid form ID", http.StatusBadRequest)
		return
	}
	
	currentStep, err := strconv.Atoi(currentStepStr)
	if err != nil || currentStep < 1 {
		http.Error(w, "Invalid step", http.StatusBadRequest)
		return
	}
	
	// Get form
	var form database.Form
	result := database.DB.First(&form, formID)
	if result.Error != nil {
		http.Error(w, "Form not found", http.StatusNotFound)
		return
	}
	
	// Get submission
	var submission database.Submission
	result = database.DB.First(&submission, submissionID)
	if result.Error != nil {
		http.Error(w, "Submission not found", http.StatusNotFound)
		return
	}
	
	// Get fields for the current step
	var fields []database.FormField
	result = database.DB.Where("form_id = ? AND step = ?", formID, currentStep).
		Order("field_order").Find(&fields)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		http.Error(w, "Failed to fetch form fields", http.StatusInternalServerError)
		return
	}
	
	// Process submitted fields
	for _, field := range fields {
		fieldIDStr := strconv.FormatUint(uint64(field.ID), 10)
		response := r.FormValue("field_" + fieldIDStr)
		
		// For required fields, validate response
		if field.IsRequired && response == "" {
			data := PageData{
				Title:      "Error",
				Error:      "Please fill out all required fields",
				Form:       form,
				Fields:     fields,
				Submission: submission,
			}
			RenderTemplate(w, "view_form.html", data)
			return
		}
		
		// Save response
		submissionResponse := database.SubmissionResponse{
			SubmissionID: uint(submissionID),
			FieldID:      field.ID,
			Response:     response,
		}
		
		result = database.DB.Create(&submissionResponse)
		if result.Error != nil {
			http.Error(w, "Failed to save response: "+result.Error.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	// Handle form navigation (next, prev, or complete)
	action := r.FormValue("action") // next, prev, or complete
	
	switch action {
	case "next":
		// Get the next step
		nextStep := currentStep + 1
		
		// Update submission
		submission.CurrentStep = nextStep
		result = database.DB.Save(&submission)
		if result.Error != nil {
			http.Error(w, "Failed to update submission: "+result.Error.Error(), http.StatusInternalServerError)
			return
		}
		
		// Get fields for the next step
		var nextFields []database.FormField
		result = database.DB.Where("form_id = ? AND step = ?", formID, nextStep).
			Order("field_order").Find(&nextFields)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			http.Error(w, "Failed to fetch form fields", http.StatusInternalServerError)
			return
		}
		
		// If no fields for next step, consider the form completed
		if len(nextFields) == 0 {
			// Mark submission as completed
			submission.Status = "completed"
			submission.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
			result = database.DB.Save(&submission)
			if result.Error != nil {
				http.Error(w, "Failed to update submission: "+result.Error.Error(), http.StatusInternalServerError)
				return
			}
			
			// Redirect to completion page
			http.Redirect(w, r, "/submissions/view/"+submission.SubmissionKey, http.StatusSeeOther)
			return
		}
		
		// Render the next step
		data := PageData{
			Title:      form.Title + " - Step " + strconv.Itoa(nextStep),
			Form:       form,
			Fields:     nextFields,
			Submission: submission,
		}
		
		RenderTemplate(w, "view_form.html", data)
		
	case "prev":
		// Get the previous step
		prevStep := currentStep - 1
		if prevStep < 1 {
			prevStep = 1
		}
		
		// Update submission
		submission.CurrentStep = prevStep
		result = database.DB.Save(&submission)
		if result.Error != nil {
			http.Error(w, "Failed to update submission: "+result.Error.Error(), http.StatusInternalServerError)
			return
		}
		
		// Get fields for the previous step
		var prevFields []database.FormField
		result = database.DB.Where("form_id = ? AND step = ?", formID, prevStep).
			Order("field_order").Find(&prevFields)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			http.Error(w, "Failed to fetch form fields", http.StatusInternalServerError)
			return
		}
		
		// Render the previous step
		data := PageData{
			Title:      form.Title + " - Step " + strconv.Itoa(prevStep),
			Form:       form,
			Fields:     prevFields,
			Submission: submission,
		}
		
		RenderTemplate(w, "view_form.html", data)
		
	case "complete":
		// Mark submission as completed
		submission.Status = "completed"
		submission.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
		result = database.DB.Save(&submission)
		if result.Error != nil {
			http.Error(w, "Failed to update submission: "+result.Error.Error(), http.StatusInternalServerError)
			return
		}
		
		// Redirect to completion page
		http.Redirect(w, r, "/submissions/view/"+submission.SubmissionKey, http.StatusSeeOther)
		
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

// generateSubmissionKey creates a unique key for submissions
func generateSubmissionKey() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to timestamp if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

	// Get event
	var event database.Event
	result := database.DB.First(&event, eventID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.NotFound(w, r)
		}
		
		result = database.DB.Create(&field)
		if result.Error != nil {
			http.Error(w, "Failed to add field: "+result.Error.Error(), http.StatusInternalServerError)
			return
		}

	case "update_field":
		fieldIDStr := r.FormValue("field_id")
		stepStr := r.FormValue("step")
		fieldType := r.FormValue("field_type")
		label := r.FormValue("label")
		placeholder := r.FormValue("placeholder")
		options := r.FormValue("options")
		isRequiredStr := r.FormValue("is_required")
		
		// Validate required fields
		if fieldIDStr == "" {
			http.Error(w, "Field ID is required", http.StatusBadRequest)
			return
		}
		
		// Parse field ID
		fieldID, err := strconv.ParseUint(fieldIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid field ID", http.StatusBadRequest)
			return
		}
		
		// Get field
		var field database.FormField
		result = database.DB.First(&field, fieldID)
		if result.Error != nil {
			http.Error(w, "Field not found", http.StatusNotFound)
			return
		}
		
		// Update field properties
		if stepStr != "" {
			step, err := strconv.Atoi(stepStr)
			if err == nil && step > 0 {
				field.Step = step
			}
		}
		
		if fieldType != "" {
			field.FieldType = fieldType
		}
		
		if label != "" {
			field.Label = label
		}
		
		field.Placeholder = placeholder
		field.Options = options
		field.IsRequired = isRequiredStr == "on" || isRequiredStr == "true"
		
		result = database.DB.Save(&field)
		if result.Error != nil {
			http.Error(w, "Failed to update field: "+result.Error.Error(), http.StatusInternalServerError)
			return
		}
		
	case "delete_field":
		fieldIDStr := r.FormValue("field_id")
		
		// Validate required fields
		if fieldIDStr == "" {
			http.Error(w, "Field ID is required", http.StatusBadRequest)
			return
		}
		
		// Parse field ID
		fieldID, err := strconv.ParseUint(fieldIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid field ID", http.StatusBadRequest)
			return
		}
		
		// Delete field
		result = database.DB.Delete(&database.FormField{}, fieldID)
		if result.Error != nil {
			http.Error(w, "Failed to delete field: "+result.Error.Error(), http.StatusInternalServerError)
			return
		}
		
	case "publish":
		publishStr := r.FormValue("publish")
		
		// Update publish status
		form.IsPublished = publishStr == "on" || publishStr == "true" || publishStr == "1"
		
		result = database.DB.Save(&form)
		if result.Error != nil {
			http.Error(w, "Failed to update publish status: "+result.Error.Error(), http.StatusInternalServerError)
			return
		}
		
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	} else {
			http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
		}
		return
	}

	data := PageData{
		Title: "Create New Form for " + event.Name,
		Event: event,
	}

	RenderTemplate(w, "new_form.html", data)
}

// CreateFormHandler handles the form submission to create a new form
func CreateFormHandler(w http.ResponseWriter, r *http.Request) {
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
	eventIDStr := r.FormValue("event_id")
	title := r.FormValue("title")
	isMultiStepStr := r.FormValue("is_multi_step")

	// Validate required fields
	if eventIDStr == "" || title == "" {
		http.Error(w, "Event ID and title are required", http.StatusBadRequest)
		return
	}

	// Parse event ID
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Check if event exists
	var event database.Event
	result := database.DB.First(&event, eventID)
	if result.Error != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Parse is_multi_step
	isMultiStep := isMultiStepStr == "on" || isMultiStepStr == "true"

	// Create new form
	form := database.Form{
		EventID:     uint(eventID),
		Title:       title,
		IsMultiStep: isMultiStep,
		IsPublished: false, // Forms are unpublished by default until fields are added
	}

	result = database.DB.Create(&form)
	if result.Error != nil {
		http.Error(w, "Failed to create form: "+result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to edit the new form (to add fields)
	http.Redirect(w, r, "/forms/edit/"+strconv.FormatUint(uint64(form.ID), 10), http.StatusSeeOther)
}

// EditFormHandler displays the form editor
func EditFormHandler(w http.ResponseWriter, r *http.Request) {
	// Extract form ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/forms/edit/")
	formID, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Get form
	var form database.Form
	result := database.DB.First(&form, formID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Failed to fetch form", http.StatusInternalServerError)
		}
		return
	}

	// Get form fields
	var fields []database.FormField
	result = database.DB.Where("form_id = ?", formID).Order("step, field_order").Find(&fields)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		http.Error(w, "Failed to fetch form fields", http.StatusInternalServerError)
		return
	}

	// Get event
	var event database.Event
	result = database.DB.First(&event, form.EventID)
	if result.Error != nil {
		http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:  "Edit Form: " + form.Title,
		Form:   form,
		Event:  event,
		Fields: fields,
	}

	RenderTemplate(w, "edit_form.html", data)
}

// UpdateFormHandler handles form updates, including adding/editing fields
func UpdateFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form with max memory limit for file uploads
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	formIDStr := r.FormValue("form_id")
	action := r.FormValue("action") // Can be: update_form, add_field, update_field, delete_field, publish

	// Validate required fields
	if formIDStr == "" {
		http.Error(w, "Form ID is required", http.StatusBadRequest)
		return
	}

	// Parse form ID
	formID, err := strconv.ParseUint(formIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid form ID", http.StatusBadRequest)
		return
	}

	// Check if form exists
	var form database.Form
	result := database.DB.First(&form, formID)
	if result.Error != nil {
		http.Error(w, "Form not found", http.StatusNotFound)
		return
	}

	// Process based on action
	switch action {
	case "update_form":
		title := r.FormValue("title")
		isMultiStepStr := r.FormValue("is_multi_step")

		if title != "" {
			form.Title = title
		}

		form.IsMultiStep = isMultiStepStr == "on" || isMultiStepStr == "true"

		result = database.DB.Save(&form)
		if result.Error != nil {
			http.Error(w, "Failed to update form: "+result.Error.Error(), http.StatusInternalServerError)
			return
		}

	case "add_field":
		stepStr := r.FormValue("step")
		fieldType := r.FormValue("field_type")
		label := r.FormValue("label")
		placeholder := r.FormValue("placeholder")
		options := r.FormValue("options")
		isRequiredStr := r.FormValue("is_required")

		// Validate required fields
		if stepStr == "" || fieldType == "" || label == "" {
			http.Error(w, "Step, field type, and label are required", http.StatusBadRequest)
			return
		}

		// Parse step
		step, err := strconv.Atoi(stepStr)
		if err != nil || step < 1 {
			step = 1
		}

		// Get highest field order for this step
		var maxOrder int
		database.DB.Model(&database.FormField{}).
			Where("form_id = ? AND step = ?", formID, step).
			Select("COALESCE(MAX(field_order), 0)").
			Scan(&maxOrder)

		// Create new field
		field := database.FormField{
			FormID:      uint(formID),
			Step:        step,
			FieldType:   fieldType,
			Label:       label,
			Placeholder: placeholder,
			Options:     options,
			IsRequired:  isRequiredStr == "on" || isRequiredStr == "true",
			FieldOrder:  maxOrder + 1,
		}
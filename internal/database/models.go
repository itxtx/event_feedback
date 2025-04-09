package database

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// Event represents an event for which feedback can be collected
type Event struct {
	gorm.Model
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Date        time.Time `gorm:"not null" json:"date"`
	Forms       []Form    `gorm:"foreignKey:EventID" json:"forms,omitempty"`
}

// TableName specifies the table name for Event
func (Event) TableName() string {
	return "events"
}

// Form represents a feedback form for an event
type Form struct {
	gorm.Model
	EventID     uint        `gorm:"index;not null" json:"event_id"`
	Title       string      `gorm:"not null" json:"title"`
	IsMultiStep bool        `gorm:"default:false" json:"is_multi_step"`
	IsPublished bool        `gorm:"default:false" json:"is_published"`
	Event       Event       `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Fields      []FormField `gorm:"foreignKey:FormID" json:"fields,omitempty"`
}

// TableName specifies the table name for Form
func (Form) TableName() string {
	return "forms"
}

// FormField represents a field in a form
type FormField struct {
	gorm.Model
	FormID      uint   `gorm:"index;not null" json:"form_id"`
	Step        int    `gorm:"default:1" json:"step"`
	FieldType   string `gorm:"not null" json:"field_type"` // text, textarea, select, radio, checkbox, etc.
	Label       string `gorm:"not null" json:"label"`
	Placeholder string `json:"placeholder"`
	Options     string `json:"options"` // JSON string for options
	IsRequired  bool   `gorm:"default:false" json:"is_required"`
	FieldOrder  int    `gorm:"not null" json:"field_order"`
	Form        Form   `gorm:"foreignKey:FormID" json:"form,omitempty"`
}

// TableName specifies the table name for FormField
func (FormField) TableName() string {
	return "form_fields"
}

// Submission represents a form submission
type Submission struct {
	gorm.Model
	FormID        uint                 `gorm:"index;not null" json:"form_id"`
	SubmissionKey string               `gorm:"uniqueIndex;not null" json:"submission_key"`
	Status        string               `gorm:"default:in_progress" json:"status"` // in_progress, completed
	CurrentStep   int                  `gorm:"default:1" json:"current_step"`
	CompletedAt   sql.NullTime         `json:"completed_at"`
	Form          Form                 `gorm:"foreignKey:FormID" json:"form,omitempty"`
	Responses     []SubmissionResponse `gorm:"foreignKey:SubmissionID" json:"responses,omitempty"`
}

// TableName specifies the table name for Submission
func (Submission) TableName() string {
	return "submissions"
}

// SubmissionResponse represents a response to a specific form field
type SubmissionResponse struct {
	gorm.Model
	SubmissionID uint       `gorm:"index;not null" json:"submission_id"`
	FieldID      uint       `gorm:"index;not null" json:"field_id"`
	Response     string     `json:"response"`
	Submission   Submission `gorm:"foreignKey:SubmissionID" json:"submission,omitempty"`
	Field        FormField  `gorm:"foreignKey:FieldID" json:"field,omitempty"`
}

// TableName specifies the table name for SubmissionResponse
func (SubmissionResponse) TableName() string {
	return "submission_responses"
}

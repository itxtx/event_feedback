package database

import (
	"database/sql"
	"time"
)

// Event represents an event for which feedback can be collected
type Event struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Date        time.Time `gorm:"not null" json:"date"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	Forms       []Form    `gorm:"foreignKey:EventID" json:"forms,omitempty"`
}

// Form represents a feedback form for an event
type Form struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	EventID     uint        `gorm:"index" json:"event_id"`
	Title       string      `gorm:"not null" json:"title"`
	IsMultiStep bool        `gorm:"default:false" json:"is_multi_step"`
	IsPublished bool        `gorm:"default:false" json:"is_published"`
	CreatedAt   time.Time   `gorm:"autoCreateTime" json:"created_at"`
	Event       Event       `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Fields      []FormField `gorm:"foreignKey:FormID" json:"fields,omitempty"`
}

// FormField represents a field in a form
type FormField struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FormID      uint      `gorm:"index" json:"form_id"`
	Step        int       `gorm:"default:1" json:"step"`
	FieldType   string    `gorm:"not null" json:"field_type"` // text, textarea, select, radio, checkbox, etc.
	Label       string    `gorm:"not null" json:"label"`
	Placeholder string    `json:"placeholder"`
	Options     string    `json:"options"` // JSON string for options
	IsRequired  bool      `gorm:"default:false" json:"is_required"`
	FieldOrder  int       `gorm:"not null" json:"field_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	Form        Form      `gorm:"foreignKey:FormID" json:"form,omitempty"`
}

// Submission represents a form submission
type Submission struct {
	ID            uint                 `gorm:"primaryKey" json:"id"`
	FormID        uint                 `gorm:"index" json:"form_id"`
	SubmissionKey string               `gorm:"uniqueIndex;not null" json:"submission_key"`
	Status        string               `gorm:"default:in_progress" json:"status"` // in_progress, completed
	CurrentStep   int                  `gorm:"default:1" json:"current_step"`
	CreatedAt     time.Time            `gorm:"autoCreateTime" json:"created_at"`
	CompletedAt   sql.NullTime         `json:"completed_at"`
	Form          Form                 `gorm:"foreignKey:FormID" json:"form,omitempty"`
	Responses     []SubmissionResponse `gorm:"foreignKey:SubmissionID" json:"responses,omitempty"`
}

// SubmissionResponse represents a response to a specific form field
type SubmissionResponse struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	SubmissionID uint       `gorm:"index" json:"submission_id"`
	FieldID      uint       `gorm:"index" json:"field_id"`
	Response     string     `json:"response"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	Submission   Submission `gorm:"foreignKey:SubmissionID" json:"submission,omitempty"`
	Field        FormField  `gorm:"foreignKey:FieldID" json:"field,omitempty"`
}

package utils

import (
	"html/template"
	"strings"
	"time"

	"github.com/yourusername/event-feedback/internal/database"
)

// TemplateFuncs returns a map of custom functions for templates
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// Returns the current time
		"now": time.Now,

		// Creates a sequence of integers from start to end
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i < end; i++ {
				result = append(result, i)
			}
			return result
		},

		// Adds two integers
		"add": func(a, b int) int {
			return a + b
		},

		// Returns the maximum step number from a slice of form fields
		"maxStep": func(fields []database.FormField) int {
			maxStep := 1
			for _, field := range fields {
				if field.Step > maxStep {
					maxStep = field.Step
				}
			}
			return maxStep
		},

		// Returns the total number of steps in a form
		"totalSteps": func(fields []database.FormField) int {
			maxStep := 1
			for _, field := range fields {
				if field.Step > maxStep {
					maxStep = field.Step
				}
			}
			return maxStep
		},

		// Splits options string into a slice of options
		"splitOptions": func(options string) []string {
			if options == "" {
				return []string{}
			}
			return strings.Split(options, "\n")
		},

		// Splits a comma-separated string of values into a slice
		"splitValues": func(values string) []string {
			if values == "" {
				return []string{}
			}
			return strings.Split(values, ",")
		},

		// Checks if a slice contains a string
		"contains": func(slice []string, item string) bool {
			for _, s := range slice {
				if s == item {
					return true
				}
			}
			return false
		},
	}
}

package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExportSchema defines the structure for CSV export configuration
// This contains the actual column definitions that should be exported
type ExportSchema struct {
	// JSON Schema reference for validation (optional)
	Schema      string `json:"$schema,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	
	// Actual column definitions that will be exported
	Columns []ExportColumn `json:"columns"`
	
	// Metadata about the schema
	SchemaVersion string `json:"schemaVersion,omitempty"`
	CreatedBy     string `json:"createdBy,omitempty"`
}

// ExportColumn represents a single column from the parsed schema
type ExportColumn struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Default     string `json:"default"`
	Type        string `json:"type"`
}

// NewEmptyExportSchema creates a new export schema with empty columns array
// This creates a minimal export configuration that users can customize
func NewEmptyExportSchema() *ExportSchema {
	return &ExportSchema{
		Schema:      "https://json-schema.org/draft/2020-12/schema",
		Title:       "TimeTracker Export Configuration",
		Description: "Configuration for TimeTracker CSV export columns",
		Type:        "object",
		Columns:     []ExportColumn{}, // Empty columns array
		SchemaVersion: "1.0",
		CreatedBy:     "TimeTracker",
	}
}

// NewDefaultExportSchema creates a new export schema with default columns
// This creates an actual export configuration with common column definitions
func NewDefaultExportSchema() *ExportSchema {
	return &ExportSchema{
		Schema:      "https://json-schema.org/draft/2020-12/schema",
		Title:       "TimeTracker Export Configuration",
		Description: "Configuration for TimeTracker CSV export columns",
		Type:        "object",
		Columns: []ExportColumn{
			{
				Name:        "project",
				DisplayName: "Project",
				Description: "Project name or identifier",
				Default:     "",
				Type:        "string",
			},
			{
				Name:        "client",
				DisplayName: "Client",
				Description: "Client or customer name",
				Default:     "",
				Type:        "string",
			},
			{
				Name:        "billable",
				DisplayName: "Billable",
				Description: "Whether the time is billable",
				Default:     "false",
				Type:        "boolean",
			},
			{
				Name:        "priority",
				DisplayName: "Priority",
				Description: "Priority level (1-5)",
				Default:     "",
				Type:        "number",
			},
			{
				Name:        "category",
				DisplayName: "Category",
				Description: "Work category (e.g., development, meeting, support)",
				Default:     "",
				Type:        "string",
			},
			{
				Name:        "tags",
				DisplayName: "Tags",
				Description: "Comma-separated tags",
				Default:     "",
				Type:        "string",
			},
		},
		SchemaVersion: "1.0",
		CreatedBy:     "TimeTracker",
	}
}

// GetExportSchemaPath returns the path to the export schema file
func (s *Storage) GetExportSchemaPath() string {
	return filepath.Join(s.baseDir, "export-schema.json")
}

// GetExportSchemaExamplePath returns the path to the export schema example file
func (s *Storage) GetExportSchemaExamplePath() string {
	return filepath.Join(s.baseDir, "export-schema-example.json")
}

// CreateExportSchemaExample creates an example export schema file with common column definitions
func (s *Storage) CreateExportSchemaExample() error {
	examplePath := s.GetExportSchemaExamplePath()
	exampleSchema := NewDefaultExportSchema()
	
	content, err := json.MarshalIndent(exampleSchema, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal export schema example: %w", err)
	}
	
	if err := os.WriteFile(examplePath, content, 0644); err != nil {
		return fmt.Errorf("could not write export schema example file: %w", err)
	}
	
	return nil
}

// LoadExportSchema loads the export schema from file
// If the file doesn't exist, it creates an empty schema and an example file
func (s *Storage) LoadExportSchema() (*ExportSchema, error) {
	schemaPath := s.GetExportSchemaPath()
	
	// Check if schema file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// Create empty schema
		schema := NewEmptyExportSchema()
		if err := s.SaveExportSchema(schema); err != nil {
			return nil, fmt.Errorf("could not save empty export schema: %w", err)
		}
		
		// Create example schema with common columns
		if err := s.CreateExportSchemaExample(); err != nil {
			// Log warning but don't fail - example is optional
			fmt.Printf("Warning: could not create export schema example: %v\n", err)
		}
		
		return schema, nil
	}
	
	// Load existing schema
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("could not read export schema file: %w", err)
	}
	
	var schema ExportSchema
	if err := json.Unmarshal(content, &schema); err != nil {
		return nil, fmt.Errorf("could not unmarshal export schema: %w", err)
	}
	
	return &schema, nil
}

// SaveExportSchema saves the export schema to file
func (s *Storage) SaveExportSchema(schema *ExportSchema) error {
	schemaPath := s.GetExportSchemaPath()
	
	content, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal export schema: %w", err)
	}
	
	if err := os.WriteFile(schemaPath, content, 0644); err != nil {
		return fmt.Errorf("could not write export schema file: %w", err)
	}
	
	return nil
}

// AddMetadataFieldToSchema adds a new metadata field to the export schema if it doesn't already exist
func (s *Storage) AddMetadataFieldToSchema(fieldName string, fieldType string) error {
	schema, err := s.LoadExportSchema()
	if err != nil {
		return fmt.Errorf("could not load export schema: %w", err)
	}

	// Check if field already exists in schema
	for _, col := range schema.Columns {
		if col.Name == fieldName {
			// Field already exists, no need to add
			return nil
		}
	}

	// Determine display name and type based on field name
	displayName := formatFieldName(fieldName)
	
	// If type is not specified, try to infer it from common patterns
	if fieldType == "" {
		fieldType = inferFieldTypeFromName(fieldName)
	}

	// Add new column to schema
	newColumn := ExportColumn{
		Name:        fieldName,
		DisplayName: displayName,
		Description: fmt.Sprintf("User-defined metadata field: %s", fieldName),
		Default:     "",
		Type:        fieldType,
	}

	schema.Columns = append(schema.Columns, newColumn)

	// Save updated schema
	if err := s.SaveExportSchema(schema); err != nil {
		return fmt.Errorf("could not save updated export schema: %w", err)
	}

	return nil
}

// ParseExportColumns extracts column definitions directly from the ExportSchema
// This returns the actual columns defined in the schema's Columns field
func (s *Storage) ParseExportColumns(schema *ExportSchema) ([]ExportColumn, error) {
	// Return the columns directly from the schema
	if schema == nil {
		return []ExportColumn{}, fmt.Errorf("schema is nil")
	}
	
	return schema.Columns, nil
}

// GetSchemaColumnNames returns the display names of all columns defined in the schema
func GetSchemaColumnNames(columns []ExportColumn) []string {
	var names []string
	for _, col := range columns {
		if col.DisplayName != "" {
			names = append(names, col.DisplayName)
		} else {
			names = append(names, col.Name)
		}
	}
	return names
}

// GetSchemaColumnMap returns a map of column names to their definitions
func GetSchemaColumnMap(columns []ExportColumn) map[string]ExportColumn {
	columnMap := make(map[string]ExportColumn)
	for _, col := range columns {
		columnMap[col.Name] = col
	}
	return columnMap
}

// formatFieldName formats a field name for display (capitalize first letter, replace underscores/dashes)
func formatFieldName(name string) string {
	if name == "" {
		return ""
	}
	
	// Replace underscores and dashes with spaces
	formatted := strings.ReplaceAll(name, "_", " ")
	formatted = strings.ReplaceAll(formatted, "-", " ")
	
	// Capitalize first letter of each word
	words := strings.Fields(formatted)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	
	return strings.Join(words, " ")
}

// inferFieldTypeFromName tries to infer the field type from the field name
func inferFieldTypeFromName(name string) string {
	nameLower := strings.ToLower(name)
	
	// Common boolean field names
	if strings.Contains(nameLower, "billable") || 
	   strings.Contains(nameLower, "active") || 
	   strings.Contains(nameLower, "completed") ||
	   strings.Contains(nameLower, "approved") ||
	   strings.Contains(nameLower, "paid") ||
	   strings.Contains(nameLower, "flag") ||
	   strings.Contains(nameLower, "is") ||
	   strings.Contains(nameLower, "has") ||
	   strings.Contains(nameLower, "can") ||
	   strings.Contains(nameLower, "should") {
		return "boolean"
	}
	
	// Common numeric field names
	if strings.Contains(nameLower, "priority") || 
	   strings.Contains(nameLower, "hours") || 
	   strings.Contains(nameLower, "duration") ||
	   strings.Contains(nameLower, "time") || 
	   strings.Contains(nameLower, "count") ||
	   strings.Contains(nameLower, "number") ||
	   strings.Contains(nameLower, "amount") ||
	   strings.Contains(nameLower, "quantity") ||
	   strings.Contains(nameLower, "size") ||
	   strings.Contains(nameLower, "rate") ||
	   strings.Contains(nameLower, "cost") ||
	   strings.Contains(nameLower, "price") {
		return "number"
	}
	
	// Default to string
	return "string"
}
package apigen

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// APIGenerator handles the generation of REST APIs from GORM models
type APIGenerator struct {
	DB              *gorm.DB
	Router          *gin.Engine
	Models          map[string]ModelInfo
	RegisteredPaths map[string]bool // Track registered paths to avoid duplicates
}

// ModelInfo stores metadata about a model
type ModelInfo struct {
	Type         reflect.Type
	Fields       []FieldInfo
	ForeignKeys  []ForeignKeyInfo
	ResourceName string
	PluralName   string
}

// FieldInfo stores metadata about a model field
type FieldInfo struct {
	Name      string
	JSONName  string
	Type      reflect.Type
	IsID      bool
	OmitEmpty bool
}

// ForeignKeyInfo stores metadata about a foreign key relationship
type ForeignKeyInfo struct {
	FieldName      string
	RelatedModel   string
	RelatedField   string
	RelationshipID string
}

// New creates a new APIGenerator instance
func New(db *gorm.DB, router *gin.Engine) *APIGenerator {
	return &APIGenerator{
		DB:              db,
		Router:          router,
		Models:          make(map[string]ModelInfo),
		RegisteredPaths: make(map[string]bool),
	}
}

// RegisterModel registers a GORM model with the API generator
func (g *APIGenerator) RegisterModel(model any, resourceName string) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct, got %s", modelType.Kind())
	}

	// If resourceName is not provided, derive it from the model name
	if resourceName == "" {
		resourceName = toSnakeCase(modelType.Name())
	}

	pluralName := pluralize(resourceName)

	modelInfo := ModelInfo{
		Type:         modelType,
		ResourceName: resourceName,
		PluralName:   pluralName,
	}

	// Process fields
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		jsonName := strings.Split(jsonTag, ",")[0]
		omitEmpty := strings.Contains(jsonTag, "omitempty")

		fieldInfo := FieldInfo{
			Name:      field.Name,
			JSONName:  jsonName,
			Type:      field.Type,
			IsID:      field.Name == "ID" || strings.HasSuffix(field.Name, "ID"),
			OmitEmpty: omitEmpty,
		}

		modelInfo.Fields = append(modelInfo.Fields, fieldInfo)

		// Check for foreign key relationships
		if field.Type.Kind() == reflect.Struct && !isBasicType(field.Type) {
			// This could be a foreign key relationship
			relatedModel := field.Type.Name()
			fkInfo := ForeignKeyInfo{
				FieldName:    field.Name,
				RelatedModel: relatedModel,
				RelatedField: "ID", // Assuming standard GORM convention
			}
			modelInfo.ForeignKeys = append(modelInfo.ForeignKeys, fkInfo)
		}

		// Check for foreign key ID fields
		if strings.HasSuffix(field.Name, "ID") && field.Type.Kind() == reflect.Uint {
			relatedModel := strings.TrimSuffix(field.Name, "ID")
			fkInfo := ForeignKeyInfo{
				FieldName:      field.Name,
				RelatedModel:   relatedModel,
				RelationshipID: field.Name,
			}
			modelInfo.ForeignKeys = append(modelInfo.ForeignKeys, fkInfo)
		}
	}

	g.Models[modelType.Name()] = modelInfo
	return nil
}

// GenerateAPI generates REST API endpoints for all registered models
func (g *APIGenerator) GenerateAPI() {
	for _, modelInfo := range g.Models {
		g.generateModelAPI(modelInfo)
	}
}

// generateModelAPI generates REST API endpoints for a specific model
func (g *APIGenerator) generateModelAPI(modelInfo ModelInfo) {
	basePath := fmt.Sprintf("/api/%s", modelInfo.PluralName)

	// Register routes
	g.Router.GET(basePath, g.listHandler(modelInfo))
	g.Router.GET(fmt.Sprintf("%s/:id", basePath), g.getHandler(modelInfo))
	g.Router.POST(basePath, g.createHandler(modelInfo))
	g.Router.PUT(fmt.Sprintf("%s/:id", basePath), g.updateHandler(modelInfo))
	g.Router.DELETE(fmt.Sprintf("%s/:id", basePath), g.deleteHandler(modelInfo))

	// Generate foreign key relationship endpoints
	for _, fk := range modelInfo.ForeignKeys {
		if fk.RelatedModel != "" {
			relatedPath := fmt.Sprintf("%s/:id/%s", basePath, toSnakeCase(fk.RelatedModel))
			
			// Check if this path has already been registered
			if !g.RegisteredPaths[relatedPath] {
				g.Router.GET(relatedPath, g.relatedHandler(modelInfo, fk))
				g.RegisteredPaths[relatedPath] = true
			}
		}
	}
}

// Helper functions for converting between naming conventions
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func toCamelCase(s string) string {
	var result strings.Builder
	nextUpper := false
	for _, r := range s {
		if r == '_' {
			nextUpper = true
		} else if nextUpper {
			result.WriteRune(unicode.ToUpper(r))
			nextUpper = false
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func pluralize(s string) string {
	// Simple pluralization rules
	if strings.HasSuffix(s, "y") {
		return strings.TrimSuffix(s, "y") + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
		strings.HasSuffix(s, "z") || strings.HasSuffix(s, "ch") ||
		strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}

func isBasicType(t reflect.Type) bool {
	// Check for time.Time type
	if t.String() == "time.Time" {
		return true
	}
	
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	}
	return false
}

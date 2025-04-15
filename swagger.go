package apigen

import (
	"fmt"
	"reflect"
	"strings"
)

// SwaggerGenerator generates Swagger documentation for the API
type SwaggerGenerator struct {
	Models map[string]ModelInfo
}

// NewSwaggerGenerator creates a new SwaggerGenerator
func NewSwaggerGenerator(models map[string]ModelInfo) *SwaggerGenerator {
	return &SwaggerGenerator{
		Models: models,
	}
}

// GenerateModelDefinitions generates Swagger model definitions for all registered models
func (g *SwaggerGenerator) GenerateModelDefinitions() map[string]any {
	definitions := make(map[string]any)

	for modelName, modelInfo := range g.Models {
		definitions[modelName] = g.generateModelDefinition(modelInfo)
	}

	return definitions
}

// generateModelDefinition generates a Swagger model definition for a specific model
func (g *SwaggerGenerator) generateModelDefinition(modelInfo ModelInfo) map[string]any {
	properties := make(map[string]any)
	required := []string{}

	for _, field := range modelInfo.Fields {
		// Skip fields that should be omitted
		if field.JSONName == "-" {
			continue
		}

		// Add the field to the properties
		properties[field.JSONName] = g.getSwaggerType(field.Type)

		// Add required fields
		if !field.OmitEmpty {
			required = append(required, field.JSONName)
		}
	}

	definition := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		definition["required"] = required
	}

	return definition
}

// GenerateRequestBody generates a Swagger request body for a model
func (g *SwaggerGenerator) GenerateRequestBody(modelInfo ModelInfo, isCreate bool) map[string]any {
	properties := make(map[string]any)
	required := []string{}

	for _, field := range modelInfo.Fields {
		// Skip fields that should be omitted or ID fields for create requests
		if field.JSONName == "-" || (isCreate && field.IsID) {
			continue
		}

		// Add the field to the properties
		properties[field.JSONName] = g.getSwaggerType(field.Type)

		// Add required fields
		if !field.OmitEmpty {
			required = append(required, field.JSONName)
		}
	}

	definition := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		definition["required"] = required
	}

	return definition
}

// GenerateResponseBody generates a Swagger response body for a model
func (g *SwaggerGenerator) GenerateResponseBody(modelInfo ModelInfo) map[string]any {
	properties := make(map[string]any)

	for _, field := range modelInfo.Fields {
		// Skip fields that should be omitted
		if field.JSONName == "-" {
			continue
		}

		// Add the field to the properties
		properties[field.JSONName] = g.getSwaggerType(field.Type)
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
	}
}

// getSwaggerType converts a Go type to a Swagger type
func (g *SwaggerGenerator) getSwaggerType(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.Bool:
		return map[string]any{
			"type": "boolean",
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{
			"type":   "integer",
			"format": g.getIntegerFormat(t),
		}
	case reflect.Float32, reflect.Float64:
		return map[string]any{
			"type":   "number",
			"format": g.getFloatFormat(t),
		}
	case reflect.String:
		return map[string]any{
			"type": "string",
		}
	case reflect.Struct:
		// Check if this is a registered model
		if modelInfo, exists := g.Models[t.Name()]; exists {
			return map[string]any{
				"$ref": fmt.Sprintf("#/definitions/%s", modelInfo.Type.Name()),
			}
		}

		// Handle time.Time
		if t.String() == "time.Time" {
			return map[string]any{
				"type":   "string",
				"format": "date-time",
			}
		}

		// For other structs, create an inline definition
		properties := make(map[string]any)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			jsonName := strings.Split(jsonTag, ",")[0]
			properties[jsonName] = g.getSwaggerType(field.Type)
		}

		return map[string]any{
			"type":       "object",
			"properties": properties,
		}
	case reflect.Slice, reflect.Array:
		return map[string]any{
			"type":  "array",
			"items": g.getSwaggerType(t.Elem()),
		}
	case reflect.Map:
		return map[string]any{
			"type":                 "object",
			"additionalProperties": g.getSwaggerType(t.Elem()),
		}
	case reflect.Ptr:
		return g.getSwaggerType(t.Elem())
	default:
		return map[string]any{
			"type": "string",
		}
	}
}

// getIntegerFormat returns the Swagger format for an integer type
func (g *SwaggerGenerator) getIntegerFormat(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int64, reflect.Uint64:
		return "int64"
	default:
		return "int32"
	}
}

// getFloatFormat returns the Swagger format for a float type
func (g *SwaggerGenerator) getFloatFormat(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Float64:
		return "double"
	default:
		return "float"
	}
}

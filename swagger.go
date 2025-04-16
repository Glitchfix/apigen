package apigen

import (
	"fmt"
	"reflect"
	"strings"
)

// SwaggerGenerator generates Swagger documentation for the API
type SwaggerGenerator struct {
	Models map[string]ModelInfo
	paths  map[string]any // internal storage for Swagger paths
}

// NewSwaggerGenerator creates a new SwaggerGenerator
func NewSwaggerGenerator(models map[string]ModelInfo) *SwaggerGenerator {
	return &SwaggerGenerator{
		Models: models,
		paths:  make(map[string]any),
	}
}

// BuildPathsForAllModels builds the Swagger paths for all CRUD endpoints (internal use)
func (g *SwaggerGenerator) BuildPathsForAllModels() {
	paths := make(map[string]any)
	for _, modelInfo := range g.Models {
		plural := modelInfo.PluralName
		modelName := modelInfo.Type.Name()
		// List endpoint
		paths["/api/"+plural] = map[string]any{
			"get": map[string]any{
				"summary":     "List all " + plural,
				"responses": map[string]any{
					"200": map[string]any{
						"description": "List response",
						"schema": map[string]any{
							"type":  "array",
							"items": map[string]any{"$ref": "#/definitions/" + modelName},
						},
					},
				},
			},
			"post": map[string]any{
				"summary":     "Create a new " + modelInfo.ResourceName,
				"parameters": []map[string]any{
					{
						"in":          "body",
						"name":        modelInfo.ResourceName,
						"description": "Create request",
						"required":    true,
						"schema":      g.GenerateRequestBody(modelInfo, true),
					},
				},
				"responses": map[string]any{
					"201": map[string]any{
						"description": "Created",
						"schema":      g.GenerateResponseBody(modelInfo),
					},
				},
			},
		}
		// Single instance endpoints
		paths["/api/"+plural+"/{id}"] = map[string]any{
			"get": map[string]any{
				"summary":     "Get a " + modelInfo.ResourceName,
				"parameters": []map[string]any{
					{"name": "id", "in": "path", "required": true, "type": "string"},
				},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Success",
						"schema":      g.GenerateResponseBody(modelInfo),
					},
					"404": map[string]any{"description": "Not found"},
				},
			},
			"put": map[string]any{
				"summary":     "Update a " + modelInfo.ResourceName,
				"parameters": []map[string]any{
					{"name": "id", "in": "path", "required": true, "type": "string"},
					{
						"in":          "body",
						"name":        modelInfo.ResourceName,
						"description": "Update request",
						"required":    true,
						"schema":      g.GenerateRequestBody(modelInfo, false),
					},
				},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Updated",
						"schema":      g.GenerateResponseBody(modelInfo),
					},
					"404": map[string]any{"description": "Not found"},
				},
			},
			"delete": map[string]any{
				"summary":     "Delete a " + modelInfo.ResourceName,
				"parameters": []map[string]any{
					{"name": "id", "in": "path", "required": true, "type": "string"},
				},
				"responses": map[string]any{
					"204": map[string]any{"description": "Deleted"},
					"404": map[string]any{"description": "Not found"},
				},
			},
		}
		// Foreign key relationships
		for _, fk := range modelInfo.ForeignKeys {
			if fk.RelatedModel != "" {
				relatedPath := fmt.Sprintf("/api/%s/{id}/%s", plural, toSnakeCase(fk.RelatedModel))
				paths[relatedPath] = map[string]any{
					"get": map[string]any{
						"summary": fmt.Sprintf("Get related %s for %s", fk.RelatedModel, modelInfo.ResourceName),
						"parameters": []map[string]any{
							{"name": "id", "in": "path", "required": true, "type": "string"},
						},
						"responses": map[string]any{
							"200": map[string]any{"description": "List response"},
						},
					},
				}
			}
		}
	}
	g.paths = paths
}

// GenerateAllPaths returns the internally built paths map
func (g *SwaggerGenerator) GenerateAllPaths() map[string]any {
	return g.paths
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

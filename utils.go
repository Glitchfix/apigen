package apigen

import (
	"fmt"
	"reflect"
	"strings"
)

// ModelAnalyzer analyzes GORM models and extracts metadata
type ModelAnalyzer struct{}

// NewModelAnalyzer creates a new ModelAnalyzer
func NewModelAnalyzer() *ModelAnalyzer {
	return &ModelAnalyzer{}
}

// AnalyzeModel analyzes a GORM model and returns its metadata
func (a *ModelAnalyzer) AnalyzeModel(model any) (ModelInfo, error) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		return ModelInfo{}, fmt.Errorf("model must be a struct, got %s", modelType.Kind())
	}

	resourceName := toSnakeCase(modelType.Name())
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

	return modelInfo, nil
}

// GenerateRequestStruct generates a request struct for a model
func (a *ModelAnalyzer) GenerateRequestStruct(modelInfo ModelInfo, isCreate bool) (string, error) {
	var builder strings.Builder

	// Generate struct name
	structName := fmt.Sprintf("%s%sRequest", modelInfo.Type.Name(), getOperationName(isCreate))
	builder.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	// Add fields
	for _, field := range modelInfo.Fields {
		// Skip ID fields for create requests
		if isCreate && field.IsID && field.Name == "ID" {
			continue
		}

		// Get the field type name
		typeName := getTypeName(field.Type)

		// Add the field to the struct
		builder.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", field.Name, typeName, field.JSONName))
	}

	builder.WriteString("}\n")
	return builder.String(), nil
}

// GenerateResponseStruct generates a response struct for a model
func (a *ModelAnalyzer) GenerateResponseStruct(modelInfo ModelInfo) (string, error) {
	var builder strings.Builder

	// Generate struct name
	structName := fmt.Sprintf("%sResponse", modelInfo.Type.Name())
	builder.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	// Add fields
	for _, field := range modelInfo.Fields {
		// Get the field type name
		typeName := getTypeName(field.Type)

		// Add the field to the struct
		builder.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", field.Name, typeName, field.JSONName))
	}

	builder.WriteString("}\n")
	return builder.String(), nil
}

// GenerateListResponseStruct generates a list response struct for a model
func (a *ModelAnalyzer) GenerateListResponseStruct(modelInfo ModelInfo) (string, error) {
	var builder strings.Builder

	// Generate struct name
	structName := fmt.Sprintf("%sListResponse", modelInfo.Type.Name())
	responseTypeName := fmt.Sprintf("%sResponse", modelInfo.Type.Name())

	builder.WriteString(fmt.Sprintf("type %s struct {\n", structName))
	builder.WriteString(fmt.Sprintf("\tItems []%s `json:\"items\"`\n", responseTypeName))
	builder.WriteString(fmt.Sprintf("\tTotal int64 `json:\"total\"`\n"))
	builder.WriteString("}\n")

	return builder.String(), nil
}

// Helper functions

// getOperationName returns the operation name based on the operation type
func getOperationName(isCreate bool) string {
	if isCreate {
		return "Create"
	}
	return "Update"
}

// getTypeName returns the type name for a reflect.Type
func getTypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int:
		return "int"
	case reflect.Int8:
		return "int8"
	case reflect.Int16:
		return "int16"
	case reflect.Int32:
		return "int32"
	case reflect.Int64:
		return "int64"
	case reflect.Uint:
		return "uint"
	case reflect.Uint8:
		return "uint8"
	case reflect.Uint16:
		return "uint16"
	case reflect.Uint32:
		return "uint32"
	case reflect.Uint64:
		return "uint64"
	case reflect.Float32:
		return "float32"
	case reflect.Float64:
		return "float64"
	case reflect.String:
		return "string"
	case reflect.Struct:
		if t.String() == "time.Time" {
			return "time.Time"
		}
		return t.Name()
	case reflect.Slice, reflect.Array:
		return "[]" + getTypeName(t.Elem())
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", getTypeName(t.Key()), getTypeName(t.Elem()))
	case reflect.Ptr:
		return "*" + getTypeName(t.Elem())
	default:
		return "any"
	}
}

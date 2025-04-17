package apigen

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// listHandler returns a handler function for listing all instances of a model
// @Summary List all instances of a model
// @Description Get all instances of a model
// @Tags API
// @Produce json
// @Success 200 {array} any
// @Router /api/{model} [get]
func (g *APIGenerator) listHandler(modelInfo ModelInfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a slice to hold the results
		sliceType := reflect.SliceOf(modelInfo.Type)
		results := reflect.New(sliceType).Interface()

		// Query the database
		if err := g.DB.Find(results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the results
		c.JSON(http.StatusOK, results)
	}
}

// getHandler returns a handler function for getting a single instance of a model by ID
// @Summary Get a model instance by ID
// @Description Get a single instance of a model by ID
// @Tags API
// @Produce json
// @Param id path string true "ID of the model instance"
// @Success 200 {object} any
// @Failure 404 {object} map[string]string
// @Router /api/{model}/{id} [get]
func (g *APIGenerator) getHandler(modelInfo ModelInfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}

		// Create a new instance of the model
		instance := reflect.New(modelInfo.Type).Interface()

		// Query the database
		idField, _ := modelInfo.Type.FieldByName("ID")
		if idField.Type.Kind() == reflect.String {
			if err := g.DB.Where("id = ?", id).First(instance).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			if err := g.DB.First(instance, id).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// Return the result
		c.JSON(http.StatusOK, instance)
	}
}

// createHandler returns a handler function for creating a new instance of a model
// @Summary Create a new model instance
// @Description Create a new instance of a model
// @Tags API
// @Accept json
// @Produce json
// @Param model body any true "Model instance"
// @Success 201 {object} any
// @Failure 400 {object} map[string]string
// @Router /api/{model} [post]
func (g *APIGenerator) createHandler(modelInfo ModelInfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new instance of the model
		instance := reflect.New(modelInfo.Type).Interface()

		// Bind the request body to the model
		if err := c.ShouldBindJSON(instance); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create the record in the database
		if err := g.DB.Create(instance).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the created instance
		c.JSON(http.StatusCreated, instance)
	}
}

// updateHandler returns a handler function for updating an instance of a model
// @Summary Update a model instance
// @Description Update an instance of a model
// @Tags API
// @Accept json
// @Produce json
// @Param id path string true "ID of the model instance"
// @Param model body any true "Model instance"
// @Success 200 {object} any
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/{model}/{id} [put]
func (g *APIGenerator) updateHandler(modelInfo ModelInfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}

		// Create a new instance of the model
		instance := reflect.New(modelInfo.Type).Interface()

		// First check if the record exists
		idField, _ := modelInfo.Type.FieldByName("ID")
		if idField.Type.Kind() == reflect.String {
			if err := g.DB.Where("id = ?", id).First(instance).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			if err := g.DB.First(instance, id).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// Bind the request body to the model
		if err := c.ShouldBindJSON(instance); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update the record in the database
		if err := g.DB.Save(instance).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the updated instance
		c.JSON(http.StatusOK, instance)
	}
}

// deleteHandler returns a handler function for deleting an instance of a model
// @Summary Delete a model instance
// @Description Delete an instance of a model
// @Tags API
// @Produce json
// @Param id path string true "ID of the model instance"
// @Success 204 {object} nil
// @Failure 404 {object} map[string]string
// @Router /api/{model}/{id} [delete]
func (g *APIGenerator) deleteHandler(modelInfo ModelInfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}

		// Create a new instance of the model
		instance := reflect.New(modelInfo.Type).Interface()

		// First check if the record exists
		idField, _ := modelInfo.Type.FieldByName("ID")
		if idField.Type.Kind() == reflect.String {
			if err := g.DB.Where("id = ?", id).First(instance).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			if err := g.DB.First(instance, id).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// Delete the record from the database
		if err := g.DB.Delete(instance).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return no content
		c.Status(http.StatusNoContent)
	}
}

// relatedHandler returns a handler function for getting related models
// @Summary Get related models
// @Description Get models related to the specified model
// @Tags API
// @Produce json
// @Param id path string true "ID of the parent model instance"
// @Success 200 {array} any
// @Failure 404 {object} map[string]string
// @Router /api/{model}/{id}/{related} [get]
func (g *APIGenerator) relatedHandler(modelInfo ModelInfo, fk ForeignKeyInfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}

		// Check if the parent record exists
		parentInstance := reflect.New(modelInfo.Type).Interface()
		if err := g.DB.First(parentInstance, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Parent record not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the related model info
		relatedModelInfo, exists := g.Models[fk.RelatedModel]
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Related model %s not registered", fk.RelatedModel)})
			return
		}

		// Create a slice to hold the results
		sliceType := reflect.SliceOf(relatedModelInfo.Type)
		results := reflect.New(sliceType).Interface()

		// Query the database for related records
		query := g.DB
		if fk.RelationshipID != "" {
			// If we have a direct foreign key ID field
			idVal, err := strconv.ParseUint(id, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
				return
			}
			query = query.Where(fk.RelationshipID, idVal)
		} else {
			// Otherwise, use the relationship name
			query = query.Where(fmt.Sprintf("%sID = ?", modelInfo.ResourceName), id)
		}

		if err := query.Find(results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the results
		c.JSON(http.StatusOK, results)
	}
}

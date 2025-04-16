package main

import (
	"github.com/Glitchfix/apigen"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func main() {
	// Initialize GORM
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	db.AutoMigrate(&User{})

	// Initialize Gin router
	router := gin.Default()

	// Register Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger.json"),
	))

	// Create API generator and register model
	apiGen := apigen.New(db, router)
	apiGen.RegisterModel(User{}, "user")
	apiGen.GenerateAPI("Minimal API", "1.0.0")

	router.Run(":8080")
}

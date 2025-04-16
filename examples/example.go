package main

import (
	"github.com/Glitchfix/apigen"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	// Swagger UI modules
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	// If you generate docs using swag init, import your docs package here
	// _ "your/module/path/docs"
)

// Example usage of the API generator

// User is a sample GORM model
type User struct {
	gorm.Model
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age,omitempty"`
	Posts []Post `json:"posts,omitempty" gorm:"foreignKey:UserID"`
}

// Post is a sample GORM model with a foreign key to User
type Post struct {
	gorm.Model
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	UserID  uint   `json:"user_id"`
	User    User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// SetupExampleAPI demonstrates how to use the API generator
func SetupExampleAPI(db *gorm.DB, router *gin.Engine) {
	// Register Swagger UI endpoint at /swagger/index.html
	// If you use swag CLI, make sure to generate docs and import the docs package above
	// This assumes your Swagger JSON is available at /swagger/doc.json (default for swag)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger.json"), // Serve Swagger UI using /swagger.json
	))
	// Create a new API generator
	apiGen := apigen.New(db, router)

	// Register models
	apiGen.RegisterModel(User{}, "user")
	apiGen.RegisterModel(Post{}, "post")

	// Generate API endpoints
	apiGen.GenerateAPI("Example API", "1.0.0")

	// Generate request and response structs
	analyzer := apigen.NewModelAnalyzer()
	userInfo := apiGen.Models["User"]

	// These would typically be used to generate code files
	_, _ = analyzer.GenerateRequestStruct(userInfo, true)  // Create request
	_, _ = analyzer.GenerateRequestStruct(userInfo, false) // Update request
	_, _ = analyzer.GenerateResponseStruct(userInfo)       // Response
	_, _ = analyzer.GenerateListResponseStruct(userInfo)   // List response
}

func main() {
	// Initialize GORM
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto migrate the models
	db.AutoMigrate(&User{}, &Post{})

	// Initialize Gin router
	router := gin.Default()

	// Setup API
	SetupExampleAPI(db, router)

	// Run the server
	router.Run(":8080")
}

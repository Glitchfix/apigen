package apigen

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Example usage of the API generator

// User is a sample GORM model
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	Age       int       `json:"age,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Posts     []Post    `json:"posts,omitempty" gorm:"foreignKey:UserID"`
}

// Post is a sample GORM model with a foreign key to User
type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SetupExampleAPI demonstrates how to use the API generator
func SetupExampleAPI(db *gorm.DB, router *gin.Engine) {
	// Create a new API generator
	apiGen := New(db, router)

	// Register models
	apiGen.RegisterModel(User{}, "user")
	apiGen.RegisterModel(Post{}, "post")

	// Generate API endpoints
	apiGen.GenerateAPI()

	// Generate Swagger documentation
	swaggerGen := NewSwaggerGenerator(apiGen.Models)
	_ = swaggerGen.GenerateModelDefinitions()

	// Generate request and response structs
	analyzer := NewModelAnalyzer()
	userInfo := apiGen.Models["User"]
	
	// These would typically be used to generate code files
	_, _ = analyzer.GenerateRequestStruct(userInfo, true)  // Create request
	_, _ = analyzer.GenerateRequestStruct(userInfo, false) // Update request
	_, _ = analyzer.GenerateResponseStruct(userInfo)       // Response
	_, _ = analyzer.GenerateListResponseStruct(userInfo)   // List response
}

/*
// Example of how to use the generated API in your application:

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
*/

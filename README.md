# 🚀 APIGen: Because Life's Too Short to Write CRUD Endpoints

> "In a world where developers spend countless hours writing the same REST endpoints over and over, one package emerged to save them all..."

## 🎯 TL;DR: What Is This Sorcery?

APIGen is your new best friend – a Go package that **automagically** generates RESTful API endpoints with all the CRUD operations your heart desires, straight from your GORM models. It's like having a tiny code-writing genie in your Go project!

**No more writing repetitive API code. No more copy-pasting endpoints. No more tears.**

## 🤔 Why You Need This In Your Life

Ever thought:

* "Gee, I'd love to spend my weekend writing 47 nearly identical CRUD endpoints!" (Said no one ever)
* "Documenting all these API endpoints manually sounds FUN!" (Also no one)
* "I sure hope this hand-written pagination code has no bugs!" (Narrator: It did)

If you nodded at any of these (and we know you didn't), then congratulations – you're either a masochist or haven't discovered APIGen yet!

## ✨ Features That Will Make You Say "Finally!"

- **Zero-Effort APIs**: Automatically conjures up REST API endpoints from your GORM models
- **Relationship Whisperer**: Handles foreign key relationships like they're old friends
- **Swagger Without Sweat**: Generates Swagger docs, because who has time to write those?
- **Request & Response Wizardry**: Creates all those tedious struct definitions for you
- **Modern Conveniences**: Pagination, filtering, and sorting – all ready to roll

## 🛠️ Installation: Easier Than Making Toast

The package is already included in your project. No additional installation needed.
Yes, you read that right – it's **ALREADY THERE**. Go grab a coffee instead!

If you're starting from scratch and want to add this magical time-saver to your project:

```bash
go install github.com/Glitchfix/apigen@latest
```

That's it! Five seconds of your life well spent.

## 🏄‍♂️ Usage: So Simple, Your Cat Could Do It

### Basic Usage: The 30-Second Setup

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/Glitchfix/apigen"
    "gorm.io/gorm"
)

// Step 1: Get your database and router ready
db, _ := gorm.Open(/* your database driver */)
router := gin.Default()

// Step 2: Wave the magic wand
apiGen := apigen.New(db, router)

// Step 3: Tell it about your models
apiGen.RegisterModel(User{}, "user")
apiGen.RegisterModel(Post{}, "post")

// Step 4: Let the magic happen
apiGen.GenerateAPI()

// Step 5: There is no step 5. You're done. Go home.
```

### Model Requirements: The Fine Print

Your GORM models need JSON tags (because we're not mind readers... yet):

```go
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" binding:"required"`
    Email     string    `json:"email" binding:"required,email"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Posts     []Post    `json:"posts,omitempty" gorm:"foreignKey:UserID"`
}
```

## 🔄 Relationships: It's Complicated (But We Handle It)

Our API generator detects those spicy foreign key relationships:

```go
type Post struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title" binding:"required"`
    Content   string    `json:"content" binding:"required"`
    UserID    uint      `json:"user_id"`
    User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

This creates relationship endpoints automagically:

- `GET /api/users/:id/posts` - Get all posts for a user, because they're clingy like that

## 📚 Swagger Documentation: Impress Your Team

Show off to your colleagues with auto-generated Swagger docs:

```go
// Generate Swagger docs with zero effort
swaggerGen := apigen.NewSwaggerGenerator(apiGen.Models)
definitions := swaggerGen.GenerateModelDefinitions()

// Your coworkers: "You wrote ALL this documentation?!"
// You: *sips coffee* "Yeah, no big deal."
```

## 🏗️ Request and Response Structs: Built While You Wait

Generate all those pesky request/response structs:

```go
analyzer := apigen.NewModelAnalyzer()
userInfo := apiGen.Models["User"]

// Generate ALL the structs!
createRequestStruct, _ := analyzer.GenerateRequestStruct(userInfo, true)
updateRequestStruct, _ := analyzer.GenerateRequestStruct(userInfo, false)
responseStruct, _ := analyzer.GenerateResponseStruct(userInfo)
listResponseStruct, _ := analyzer.GenerateListResponseStruct(userInfo)

// While your colleagues are still typing struct definitions, you're already at lunch
```

## 🛣️ Endpoints: The Promised Land

For each model, you get these beautiful endpoints (no assembly required):

- `GET /api/{models}` - List all instances (with pagination!)
- `GET /api/{models}/:id` - Get a specific instance
- `POST /api/{models}` - Create something new and exciting
- `PUT /api/{models}/:id` - Update when you made a boo-boo
- `DELETE /api/{models}/:id` - Make it disappear
- `GET /api/{models}/:id/{related}` - Explore those relationships

## 🧪 Complete Working Example: See It In Action

```go
package apigen

import (
	"time"

    "github.com/Glitchfix/apigen"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
	
	// These would typically be used to generate code files, but this is block below isnot really necessary 
	_, _ = analyzer.GenerateRequestStruct(userInfo, true)  // Create request
	_, _ = analyzer.GenerateRequestStruct(userInfo, false) // Update request
	_, _ = analyzer.GenerateResponseStruct(userInfo)       // Response
	_, _ = analyzer.GenerateListResponseStruct(userInfo)   // List response
}

// Example of how to use the generated API in your application:
func main() {
	// Initialize GORM
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database") // Panic: the only sensible reaction to database connection issues
	}

	// Auto migrate the models
	db.AutoMigrate(&User{}, &Post{})

	// Initialize Gin router
	router := gin.Default()

	// Setup API (the easy part)
	SetupExampleAPI(db, router)

	// Run the server and go home early
	router.Run(":8080")
}
```

## 🎉 Conclusion: Save Time, Have Fun, Be a Hero

APIGen lets you focus on the interesting parts of your application instead of writing boilerplate CRUD operations all day. Your productivity will skyrocket, your code will be more consistent, and your coworkers will wonder how you finish projects so quickly.

Get started now and reclaim those hours of your life you'd otherwise spend writing repetitive API code. Your future self will thank you!

## 🚀 Minimal Example

```go
package main

import (
	"github.com/Glitchfix/apigen"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type User struct {
	gorm.Model
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func main() {
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	db.AutoMigrate(&User{})

	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger.json"),
	))

	apiGen := apigen.New(db, router)
	apiGen.RegisterModel(User{}, "user")
	apiGen.GenerateAPI("Minimal API", "1.0.0")

	router.Run(":8080")
}
```

Visit [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) for beautiful, interactive docs!

## 🔗 Resources & Community

- **Star us on GitHub** - It makes our day and helps others find this time-saving tool
- **Report Bugs** - Find a bug? Let us know! (Though with auto-generated code, there are obviously none... right?)
- **Feature Requests** - Have ideas? We're all ears!

---

*Made with ❤️ by developers who got tired of writing the same CRUD endpoints over and over again.*

*Remember: A good developer is a lazy developer.*

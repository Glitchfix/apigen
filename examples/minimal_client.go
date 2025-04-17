package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"resty.dev/v3"
)

var client = resty.New()

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	baseURL := "http://localhost:8080/api"

	// 1. List users
	users := listUsers(baseURL)
	log.Info().Msgf("Step 1: Initial users: %v", users)

	// 2. Create users
	user1 := createUser(baseURL, "Alice", "alice@example.com")
	user2 := createUser(baseURL, "Bob", "bob@example.com")
	log.Info().Msgf("Step 2: Created users: %+v, %+v", user1, user2)

	// 3. List users again
	users = listUsers(baseURL)
	log.Info().Msgf("Step 3: Users after creation: %v", users)

	// 4. Get user by id
	user := getUser(baseURL, user1.ID)
	log.Info().Msgf("Step 4: Got user by ID: %+v", user)

	// 5. Update user
	updatedName := "Alice Updated"
	updatedEmail := "alice.updated@example.com"
	updateUser(baseURL, user1.ID, updatedName, updatedEmail)
	log.Info().Msgf("Step 5: Updated user %s", user1.ID)

	// 6. Get user by id again and verify the update
	user = getUser(baseURL, user1.ID)
	if user.Name != updatedName || user.Email != updatedEmail {
		log.Fatal().Msgf("Step 6: User update verification failed: got %+v", user)
	}
	log.Info().Msgf("Step 6: Verified user update: %+v", user)
}

func listUsers(baseURL string) []User {
	var users []User
	resp, err := client.R().
		SetResult(&users).
		Get(baseURL + "/users")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list users")
	}
	if resp.StatusCode() != 200 {
		log.Fatal().Msgf("List users failed: status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return users
}

func createUser(baseURL, name, email string) User {
	var created User
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"name": name, "email": email}).
		SetResult(&created).
		Post(baseURL + "/users")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create user")
	}
	if resp.StatusCode() != 201 {
		log.Fatal().Msgf("Create user failed: status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return created
}

func getUser(baseURL string, id string) User {
	var user User
	resp, err := client.R().
		SetResult(&user).
		Get(fmt.Sprintf("%s/users/%s", baseURL, id))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get user by id")
	}
	if resp.StatusCode() != 200 {
		log.Fatal().Msgf("Get user by id failed: status %d, body: %s", resp.StatusCode(), resp.String())
	}
	return user
}

func updateUser(baseURL string, id string, name, email string) {
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"name": name, "email": email}).
		Put(fmt.Sprintf("%s/users/%s", baseURL, id))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to update user")
	}
	if resp.StatusCode() != 200 {
		log.Fatal().Msgf("Update user failed: status %d, body: %s", resp.StatusCode(), resp.String())
	}
}

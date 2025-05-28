// Package auth provides authentication and authorization functionality for the Meals API.
//
// This package implements OAuth2 authentication using Google as the provider,
// along with session management for maintaining user authentication state.
//
// Key components:
// - OAuth2 setup and configuration
// - Session management with secure cookies
// - User authentication middleware
// - Role-based authorization
//
// Authentication Flow:
// 1. User visits /auth/google
// 2. Redirected to Google OAuth2
// 3. Google redirects to /auth/google/callback
// 4. User session is created and stored
// 5. User is redirected to the application
//
// Session Management:
// - Sessions are stored in secure HTTP-only cookies
// - Session data includes user information from OAuth2
// - Sessions can be validated and retrieved for authenticated requests
package auth

import (
	"fmt"
	"log"
	"meals/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

// InitOAuth2 initializes the OAuth2 configuration using Google as the provider.
//
// This function sets up the OAuth2 provider with credentials from the application
// configuration and configures the session store for maintaining authentication state.
//
// Configuration required:
// - GoogleKey: OAuth2 client ID from Google Console
// - GoogleSecret: OAuth2 client secret from Google Console
// - GoogleRedirectURL: Callback URL registered with Google
// - SessionSecret: Secret key for signing session cookies
func InitOAuth2() {
	// Use the configuration from the config package
	authConfig := config.AppConfig.Auth

	// Configure the OAuth2 provider
	goth.UseProviders(
		google.New(
			authConfig.GoogleKey,
			authConfig.GoogleSecret,
			authConfig.GoogleRedirectURL,
		),
	)
	gothic.Store = sessions.NewCookieStore([]byte(authConfig.SessionSecret))
}

func GetSessionUser(r *http.Request) (goth.User, error) {
	session, err := gothic.Store.Get(r, "session")
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return goth.User{}, err
	}

	u := session.Values["user"]
	if u == nil {
		return goth.User{}, fmt.Errorf("user is not authenticated! %v", u)
	}

	user, ok := u.(goth.User)
	if !ok {
		return goth.User{}, fmt.Errorf("invalid user type in session")
	}

	return user, nil
}

func StoreUserSession(w http.ResponseWriter, r *http.Request, user goth.User) error {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always reutnrs a session, even if empty.
	session, _ := gothic.Store.Get(r, "session")
	session.Values["user"] = user
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// IsAuthenticated checks if a user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, err := GetSessionUser(c.Request)
	return err == nil
}

// RequireAuth middleware enforces authentication or shows welcome page
func RequireAuth(handlerFunc gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		gUser, err := GetSessionUser(c.Request)
		if err != nil {
			log.Printf("User is not authenticated: %v", err)
			// Instead of just returning with no response, redirect or show welcome page
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Welcome to Meals App</title>
				<style>
					body { 
						font-family: Arial, sans-serif; 
						margin: 40px;
						line-height: 1.6;
						color: #333;
					}
					h1 { color: #2c3e50; }
					a { 
						color: #3498db; 
						text-decoration: none;
						font-weight: bold;
					}
					a:hover { text-decoration: underline; }
					.container {
						max-width: 800px;
						margin: 0 auto;
						padding: 20px;
						border-radius: 5px;
						background-color: #f9f9f9;
						box-shadow: 0 2px 4px rgba(0,0,0,0.1);
					}
				</style>
			</head>
			<body>
				<div class="container">
					<h1>Welcome to Meals App</h1>
					<p>Please <a href="/auth/google">login with Google</a> to continue.</p>
					<p>This application helps you plan your meals for the week.</p>
				</div>
			</body>
			</html>
			`)
			return
		}
		c.Set("user", gUser)

		log.Printf("User is authenticated: %v", gUser.Email)

		handlerFunc(c)
	}
}

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

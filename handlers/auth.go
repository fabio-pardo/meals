package handlers

import (
	"log"
	"meals/auth"
	"meals/models"
	"meals/store"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"gorm.io/gorm"
)



func GetAuthProviderHandler(c *gin.Context) {
	c.Request = setProviderInRequest(c.Request, c.Param("provider"))
	if gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request); err == nil {
		log.Printf("User already authenticated! %v", gothUser)
		// Redirect already authenticated users to the home page
		c.Redirect(http.StatusFound, "/")
	} else {
		// Begin the authentication process
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}
}

func GetAuthCallbackHandler(c *gin.Context) {
	c.Request = setProviderInRequest(c.Request, c.Param("provider"))
	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		log.Printf("Failed to complete authentication: %s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	
	// After successful authentication, log the attempt
	log.Printf("Successfully authenticated user: %s (%s)", gothUser.Name, gothUser.Email)

	var existingUser models.User
	result := store.DB.Where("user_id = ?", gothUser.UserID).First(&existingUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			newUser, err := models.ConvertGothUserToModelUser(&gothUser)
			if err != nil {
				RespondWithError(c, ErrorResponse{
					Status:  http.StatusInternalServerError,
					Code:    ErrInternalServer,
					Message: "Failed to convert user data",
				})
				return
			}
			if err := store.DB.Create(newUser).Error; err != nil {
				RespondWithError(c, DatabaseError("Failed to create user in database"))
				return
			}
		} else {
			RespondWithError(c, DatabaseError("Failed connection to DB"))
			return
		}
	} else {
		existingUser.UserID = gothUser.UserID
		existingUser.AccessToken = gothUser.AccessToken
		existingUser.AccessTokenSecret = gothUser.AccessTokenSecret
		existingUser.RefreshToken = gothUser.RefreshToken
		existingUser.ExpiresAt = gothUser.ExpiresAt
		// Save the changes to the database
		if err := store.DB.Save(&existingUser).Error; err != nil {
			log.Printf("Failed to update gothUser to pre-existing DB User %s", gothUser.UserID)
			RespondWithError(c, DatabaseError("Failed to update user credentials"))
			return
		}
	}

	// Store user in session cookies
	err = auth.StoreUserSession(c.Writer, c.Request, gothUser)
	if err != nil {
		log.Printf("Failed to store user session: %s", err.Error())
		RespondWithError(c, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Code:    ErrInternalServer,
			Message: "Failed to store user session",
		})
		return
	}

	// Redirect to home page after successful authentication
	c.Redirect(http.StatusFound, "/")
}

// LogoutHandler handles user logout by clearing the session
func LogoutHandler(c *gin.Context) {
	session, _ := gothic.Store.Get(c.Request, "session")
	
	// Remove user from session
	delete(session.Values, "user")
	
	// Save session
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Printf("Error saving session during logout: %v", err)
	}
	
	// Redirect to home page
	c.Redirect(http.StatusFound, "/")
}

// Helper function to inject the provider into the request context
func setProviderInRequest(req *http.Request, provider string) *http.Request {
	q := req.URL.Query()
	q.Add("provider", provider)
	req.URL.RawQuery = q.Encode()
	return req
}

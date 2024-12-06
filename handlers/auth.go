package handlers

import (
	"fmt"
	"html/template"
	"log"
	"meals/auth"
	"meals/models"
	"meals/store"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"gorm.io/gorm"
)

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`

func GetAuthProviderHandler(c *gin.Context) {
	c.Request = setProviderInRequest(c.Request, c.Param("provider"))
	if gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request); err == nil {
		log.Printf("User already authenticated! %v", gothUser)
		t, _ := template.New("foo").Parse(userTemplate)
		t.Execute(c.Writer, gothUser)
	} else {
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

	var existingUser models.User
	result := store.DB.Where("user_id = ?", gothUser.UserID).First(&existingUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			newUser, _ := models.ConvertGothUserToModelUser(&gothUser)
			store.DB.Create(newUser)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed connection to DB"})
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
			return
		}
	}

	// Store user in session cookies
	err = auth.StoreUserSession(c.Writer, c.Request, gothUser)
	if err != nil {
		log.Printf("Failed to store user session: %s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	// Display user information
	response := fmt.Sprintf("User Info: \nName: %s\nEmail: %s\n", gothUser.Name, gothUser.Email)
	c.String(http.StatusOK, response)
}

// Helper function to inject the provider into the request context
func setProviderInRequest(req *http.Request, provider string) *http.Request {
	q := req.URL.Query()
	q.Add("provider", provider)
	req.URL.RawQuery = q.Encode()
	return req
}

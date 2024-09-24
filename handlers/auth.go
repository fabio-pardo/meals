package handlers

import (
	"fmt"
	"html/template"
	"log"
	"meals/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
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
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		log.Printf("Failed to complete authentication: %s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	err = auth.StoreUserSession(c.Writer, c.Request, user)
	if err != nil {
		log.Printf("Failed to store user session: %s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	// Display user information
	response := fmt.Sprintf("User Info: \nName: %s\nEmail: %s\n", user.Name, user.Email)
	c.String(http.StatusOK, response)
}

// Helper function to inject the provider into the request context
func setProviderInRequest(req *http.Request, provider string) *http.Request {
	q := req.URL.Query()
	q.Add("provider", provider)
	req.URL.RawQuery = q.Encode()
	return req
}

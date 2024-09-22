package controllers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

var userTemplate = `
<p><a href="/auth/logout/{{.Provider}}">logout</a></p>
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

func AuthHandler(c *gin.Context) {
	c.Request = setProviderInRequest(c.Request, c.Param("provider"))
	if gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request); err == nil {
		t, _ := template.New("foo").Parse(userTemplate)
		t.Execute(c.Writer, gothUser)
	} else {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}
}

func AuthCallback(c *gin.Context) {
	c.Request = setProviderInRequest(c.Request, c.Param("provider"))
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		fmt.Fprintln(c.Writer, err)
		return
	}
	t, _ := template.New("foo").Parse(userTemplate)
	t.Execute(c.Writer, user)
}

func AuthLogout(c *gin.Context) {
	c.Request = setProviderInRequest(c.Request, c.Param("provider"))
	gothic.Logout(c.Writer, c.Request)
	c.Writer.Header().Set("Location", "/")
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}

// Helper function to inject the provider into the request context
func setProviderInRequest(req *http.Request, provider string) *http.Request {
	q := req.URL.Query()
	q.Add("provider", provider)
	req.URL.RawQuery = q.Encode()
	return req
}

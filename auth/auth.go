package auth

import (
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var (
	Store  = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	key    = "randomString"
	MaxAge = 86400 * 30
	IsProd = false
)

func InitOAuth2() {
	// Configure the OAuth2 provider
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_KEY"),
			os.Getenv("GOOGLE_SECRET"),
			os.Getenv("GOOGLE_REDIRECT_URL"),
		),
	)
	Store.Options.MaxAge = MaxAge
	Store.Options.HttpOnly = true
	Store.Options.Secure = IsProd

	gothic.Store = Store
}

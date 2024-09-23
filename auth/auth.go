package auth

import (
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
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
}

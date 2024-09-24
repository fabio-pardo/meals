package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
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
	gothic.Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
}

func GetSessionUser(r *http.Request) (goth.User, error) {
	session, err := gothic.Store.Get(r, "session")
	if err != nil {
		return goth.User{}, err
	}

	u := session.Values["user"]
	if u == nil {
		return goth.User{}, fmt.Errorf("user is not authenticated! %v", u)
	}
	return u.(goth.User), nil
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

func RequireAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSessionUser(r)
		if err != nil {
			log.Println("User is not authenticated!")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		log.Printf("user is authenticated! user: %v!", session.FirstName)

		handlerFunc(w, r)
	}
}

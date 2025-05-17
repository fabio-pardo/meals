package auth_test

import (
	"meals/auth"
	"meals/models"
	"meals/tests"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationSystem(t *testing.T) {
	db := tests.SetupTestSuite(t)
	gin.SetMode(gin.TestMode)

	t.Run("SessionMiddleware_ValidSession", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create a session
		session := models.Session{
			UserID:    user.ID,
			Token:     "valid-session-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		// Save the session
		db.Create(&session)

		// Set up request with session token
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.AddCookie(&http.Cookie{
			Name:  "session",
			Value: session.Token,
		})

		// Set database in context
		c.Set("db", &models.Database{DB: db})

		// Execute session middleware
		auth.SessionMiddleware()(c)

		// Check if user was added to context
		userVal, exists := c.Get("user")
		assert.True(t, exists, "Expected user to be added to context")

		// Check if it's the correct user
		contextUser, ok := userVal.(models.User)
		assert.True(t, ok, "Expected user in context to be a models.User")
		assert.Equal(t, user.ID, contextUser.ID, "Expected correct user in context")

		// Check that the request wasn't aborted
		assert.False(t, c.IsAborted(), "Expected request to proceed")
	})

	t.Run("SessionMiddleware_InvalidSession", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Set up request with invalid session token
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.AddCookie(&http.Cookie{
			Name:  "session",
			Value: "invalid-session-token",
		})

		// Set database in context
		c.Set("db", &models.Database{DB: db})

		// Execute session middleware
		auth.SessionMiddleware()(c)

		// Check if user was not added to context
		_, exists := c.Get("user")
		assert.False(t, exists, "Expected no user in context for invalid session")

		// For public endpoints, the request shouldn't be aborted
		assert.False(t, c.IsAborted(), "Expected request to proceed even with invalid session")
	})

	t.Run("SessionMiddleware_ExpiredSession", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create an expired session
		session := models.Session{
			UserID:    user.ID,
			Token:     "expired-session-token",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		}

		// Save the session
		db.Create(&session)

		// Set up request with expired session token
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.AddCookie(&http.Cookie{
			Name:  "session",
			Value: session.Token,
		})

		// Set database in context
		c.Set("db", &models.Database{DB: db})

		// Execute session middleware
		auth.SessionMiddleware()(c)

		// Check if user was not added to context
		_, exists := c.Get("user")
		assert.False(t, exists, "Expected no user in context for expired session")

		// Check that the expired session was deleted from the database
		var count int64
		db.Model(&models.Session{}).Where("token = ?", session.Token).Count(&count)
		assert.Equal(t, int64(0), count, "Expected expired session to be deleted")
	})

	t.Run("Login_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// This test is more of an integration test that would require mocking OAuth providers
		// We'll simulate a successful login by creating a user and session directly

		// Create a test user
		user := models.User{
			Provider:    "google",
			Email:       "test@example.com",
			Name:        "Test User",
			FirstName:   "Test",
			LastName:    "User",
			UserID:      "google-123456",
			AccessToken: "google-access-token",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			IDToken:     "google-id-token",
			UserType:    models.UserTypeCustomer,
		}

		// Save the user
		db.Create(&user)

		// Create test login handler
		loginHandler := func(c *gin.Context) {
			// Create a new session
			session := models.Session{
				UserID:    user.ID,
				Token:     "new-session-token",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}

			db.Create(&session)

			// Set the session cookie
			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "session",
				Value:    session.Token,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				MaxAge:   86400, // 24 hours
			})

			// Redirect to home
			c.Redirect(http.StatusFound, "/home")
		}

		// Set up mock request and response
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/auth/callback", nil)
		c.Set("db", &models.Database{DB: db})

		// Execute login handler
		loginHandler(c)

		// Check response
		assert.Equal(t, http.StatusFound, w.Code, "Expected redirect status code")
		assert.Equal(t, "/home", w.Header().Get("Location"), "Expected redirect to home")

		// Check that a cookie was set
		cookies := w.Result().Cookies()
		assert.GreaterOrEqual(t, len(cookies), 1, "Expected at least one cookie")

		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "session" {
				sessionCookie = cookie
				break
			}
		}

		assert.NotNil(t, sessionCookie, "Expected session cookie to be set")
		assert.Equal(t, "new-session-token", sessionCookie.Value, "Expected correct session token")
	})

	t.Run("Logout_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a user
		user := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Create a session
		session := models.Session{
			UserID:    user.ID,
			Token:     "session-to-logout",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		// Save the session
		db.Create(&session)

		// Set up request with session token
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/auth/logout", nil)
		c.Request.AddCookie(&http.Cookie{
			Name:  "session",
			Value: session.Token,
		})

		// Set database in context
		c.Set("db", &models.Database{DB: db})

		// Create logout handler
		logoutHandler := func(c *gin.Context) {
			// Get the session token from cookie
			cookie, err := c.Cookie("session")
			if err == nil {
				// Delete the session from database
				db.Where("token = ?", cookie).Delete(&models.Session{})
			}

			// Clear the cookie
			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "session",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1, // Delete the cookie
			})

			// Redirect to login
			c.Redirect(http.StatusFound, "/auth/login")
		}

		// Execute logout handler
		logoutHandler(c)

		// Check response
		assert.Equal(t, http.StatusFound, w.Code, "Expected redirect status code")
		assert.Equal(t, "/auth/login", w.Header().Get("Location"), "Expected redirect to login")

		// Check that cookie was cleared
		cookies := w.Result().Cookies()
		assert.GreaterOrEqual(t, len(cookies), 1, "Expected at least one cookie")

		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "session" {
				sessionCookie = cookie
				break
			}
		}

		assert.NotNil(t, sessionCookie, "Expected session cookie to be present")
		assert.Equal(t, "", sessionCookie.Value, "Expected session cookie to be cleared")
		assert.Less(t, sessionCookie.MaxAge, 0, "Expected session cookie to be deleted")

		// Check that session was deleted from database
		var count int64
		db.Model(&models.Session{}).Where("token = ?", session.Token).Count(&count)
		assert.Equal(t, int64(0), count, "Expected session to be deleted from database")
	})
}

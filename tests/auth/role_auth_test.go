package auth_test

import (
	"meals/auth"
	"meals/models"
	"meals/tests"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRoleBasedAuthorization(t *testing.T) {
	db := tests.SetupTestSuite(t)
	gin.SetMode(gin.TestMode)

	t.Run("RequireAdmin_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create admin user
		adminUser := tests.CreateTestUser(db, models.UserTypeAdmin)

		// Set up test context with admin user
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/admin", nil)
		c.Set("user", adminUser)

		// Apply middleware
		auth.RequireAdmin()(c)

		// Check that the request was allowed to proceed
		assert.False(t, c.IsAborted(), "Expected admin request to proceed")
	})

	t.Run("RequireAdmin_Failure_CustomerUser", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create customer user
		customerUser := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Set up test context with customer user
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/admin", nil)
		c.Set("user", customerUser)

		// Apply middleware
		auth.RequireAdmin()(c)

		// Check that the request was aborted
		assert.True(t, c.IsAborted(), "Expected customer request to be aborted")
		assert.Equal(t, http.StatusForbidden, w.Code, "Expected forbidden status code")
	})

	t.Run("RequireRole_Success_SingleRole", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create driver user
		driverUser := tests.CreateTestUser(db, models.UserTypeDriver)

		// Set up test context with driver user
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/driver", nil)
		c.Set("user", driverUser)

		// Apply middleware requiring driver role
		auth.RequireRole(models.UserTypeDriver)(c)

		// Check that the request was allowed to proceed
		assert.False(t, c.IsAborted(), "Expected driver request to proceed")
	})

	t.Run("RequireRole_Success_MultipleRoles", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create customer user
		customerUser := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Set up test context with customer user
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/user-content", nil)
		c.Set("user", customerUser)

		// Apply middleware accepting both customer and driver roles
		auth.RequireRole(models.UserTypeCustomer, models.UserTypeDriver)(c)

		// Check that the request was allowed to proceed
		assert.False(t, c.IsAborted(), "Expected customer request to proceed with multi-role middleware")
	})

	t.Run("RequireRole_Failure_WrongRole", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create customer user
		customerUser := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Set up test context with customer user
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/driver-only", nil)
		c.Set("user", customerUser)

		// Apply middleware requiring driver role
		auth.RequireRole(models.UserTypeDriver)(c)

		// Check that the request was aborted
		assert.True(t, c.IsAborted(), "Expected customer request to be aborted for driver-only endpoint")
		assert.Equal(t, http.StatusForbidden, w.Code, "Expected forbidden status code")
	})

	t.Run("RequireRole_Failure_NoUser", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Set up test context with no user
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/authenticated", nil)

		// Apply middleware requiring customer role
		auth.RequireRole(models.UserTypeCustomer)(c)

		// Check that the request was aborted
		assert.True(t, c.IsAborted(), "Expected unauthenticated request to be aborted")
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected unauthorized status code")
	})

	t.Run("RequireAny_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create customer user
		customerUser := tests.CreateTestUser(db, models.UserTypeCustomer)

		// Set up test context with customer user
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/any-user", nil)
		c.Set("user", customerUser)

		// Apply middleware requiring any authenticated user
		auth.RequireRole("")(c)

		// Check that the request was allowed to proceed
		assert.False(t, c.IsAborted(), "Expected authenticated request to proceed")
	})

	t.Run("RequireAny_Failure", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Set up test context with no user
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/any-user", nil)

		// Apply middleware requiring any authenticated user
		auth.RequireRole("")(c)

		// Check that the request was aborted
		assert.True(t, c.IsAborted(), "Expected unauthenticated request to be aborted")
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected unauthorized status code")
	})
}

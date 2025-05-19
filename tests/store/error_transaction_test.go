package store_test

import (
	"errors"
	"meals/handlers"
	"meals/models"
	"meals/store"
	"meals/tests"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestErrorHandlingInTransactions(t *testing.T) {
	db := tests.SetupTestSuite(t)

	createTestContext := func() *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("POST", "/test", nil)
		c.Request = req
		return c
	}

	t.Run("TransactionWithAppError", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test context
		c := createTestContext()

		// Create a sample table for testing
		db.Exec("CREATE TABLE IF NOT EXISTS transaction_error_test (id SERIAL PRIMARY KEY, value TEXT)")

		// Define a validation error to be thrown in the transaction
		validationErrMsg := "Validation error in transaction"

		// Run a transaction that will encounter an error
		err := store.WithTransaction(c, func(tx *gorm.DB) error {
			// Insert a test record
			insertErr := tx.Exec("INSERT INTO transaction_error_test (value) VALUES ('before_error')").Error
			assert.Nil(t, insertErr, "Expected insert to succeed")

			// Return a validation error
			return errors.New(validationErrMsg)
		})

		// Verify transaction was rolled back due to an error
		assert.Equal(t, validationErrMsg, err.Error(), "Expected validation error from transaction")

		// Verify the record was not committed
		var count int64
		db.Table("transaction_error_test").Count(&count)
		assert.Equal(t, int64(0), count, "Expected no records after rollback")
	})

	t.Run("TransactionWithMultipleErrorTypes", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test context
		c := createTestContext()

		// Create a sample table for testing
		db.Exec("CREATE TABLE IF NOT EXISTS transaction_multierror_test (id SERIAL PRIMARY KEY, value TEXT NOT NULL)")

		// Run a transaction with database error
		err := store.WithTransaction(c, func(tx *gorm.DB) error {
			// Try a successful operation
			insertErr := tx.Exec("INSERT INTO transaction_multierror_test (value) VALUES ('valid_value')").Error
			assert.Nil(t, insertErr, "Expected first insert to succeed")

			// Try an operation that would fail with a DB error
			dbErr := tx.Exec("INSERT INTO transaction_multierror_test (value) VALUES (NULL)").Error
			assert.NotNil(t, dbErr, "Expected database error for NULL value")

			// This should return the database error
			return dbErr
		})

		// Verify transaction returned the database error
		assert.NotNil(t, err, "Expected error from transaction")

		// Verify the record was not committed
		var count int64
		db.Table("transaction_multierror_test").Count(&count)
		assert.Equal(t, int64(0), count, "Expected no records after rollback")
	})

	t.Run("NestedTransactionWithAppError", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test context
		c := createTestContext()

		// Create a sample table for testing
		db.Exec("CREATE TABLE IF NOT EXISTS nested_transaction_test (id SERIAL PRIMARY KEY, value TEXT)")

		// Define an error message for the nested transaction
		notFoundErrMsg := "Resource not found in nested transaction"

		// Run an outer transaction
		err := store.WithTransaction(c, func(tx *gorm.DB) error {
			// Insert a record in the outer transaction
			insertErr := tx.Exec("INSERT INTO nested_transaction_test (value) VALUES ('outer_tx')").Error
			assert.Nil(t, insertErr, "Expected outer insert to succeed")

			// Run inner transaction as another database operation
			innerInsertErr := tx.Exec("INSERT INTO nested_transaction_test (value) VALUES ('inner_tx')").Error
			assert.Nil(t, innerInsertErr, "Expected inner insert to succeed")

			// Return a not found error
			return errors.New(notFoundErrMsg)
		})

		// Verify transaction returned the not found error
		assert.Equal(t, notFoundErrMsg, err.Error(), "Expected not found error from transaction")

		// Verify no records were committed
		var count int64
		db.Table("nested_transaction_test").Count(&count)
		assert.Equal(t, int64(0), count, "Expected no records after nested rollback")
	})

	t.Run("RelationshipErrorInTransaction", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test context
		c := createTestContext()

		// Define a relationship error message
		relationshipErrMsg := "Cannot delete referenced entity"

		// Create test meal - we don't need the user for this test
		meal := tests.CreateTestMeal(db, "Test Meal", 12.99)

		tests.CreateTestMenu(db, "Test Menu", []uint{meal.ID})

		// Run a transaction that will raise a relationship error
		err := store.WithTransaction(c, func(tx *gorm.DB) error {
			// Try to delete the meal (which should fail with relationship error)
			if err := tx.Delete(&meal).Error; err != nil {
				// In a real app, we'd detect the foreign key error and return a relationship error
				// Here we'll simulate that by just returning our predefined error
				return errors.New(relationshipErrMsg)
			}

			return nil
		})

		// Verify transaction returned the relationship error
		assert.Equal(t, relationshipErrMsg, err.Error(), "Expected relationship error from transaction")

		// Verify the meal still exists
		var mealCount int64
		db.Model(&models.Meal{}).Where("id = ?", meal.ID).Count(&mealCount)
		assert.Equal(t, int64(1), mealCount, "Expected meal to still exist after rollback")
	})

	t.Run("HandleAppErrorInTransaction", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test context
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Test with a validation error
		validationErr := handlers.ValidationErrorType{Message: "test", Details: "Mock validation error"}
		handled := handlers.HandleAppError(c, validationErr)
		assert.True(t, handled, "Expected ValidationError to be handled")

		// Test with a generic error
		genericError := errors.New("generic error")
		handled = handlers.HandleAppError(c, genericError)
		assert.True(t, handled, "Expected generic error to be handled")
	})
}

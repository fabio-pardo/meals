package store_test

import (
	"errors"
	"meals/store"
	"meals/tests"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestBasicTransactions tests the basic functionality of the transaction system
func TestBasicTransactions(t *testing.T) {
	db := tests.SetupTestSuite(t)
	
	t.Run("SuccessfulTransaction", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a test context
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		
		// Create a sample table for testing
		db.Exec("CREATE TABLE IF NOT EXISTS transaction_success_test (id SERIAL PRIMARY KEY, value TEXT)")
		
		// Run a transaction that succeeds
		err := store.WithTransaction(c, func(tx *gorm.DB) error {
			// Insert a test record
			return tx.Exec("INSERT INTO transaction_success_test (value) VALUES ('success')").Error
		})
		
		// Verify transaction succeeded
		assert.Nil(t, err, "Expected no error from successful transaction")
		
		// Verify the record was committed
		var count int64
		db.Table("transaction_success_test").Count(&count)
		assert.Equal(t, int64(1), count, "Expected record to be committed after successful transaction")
	})
	
	t.Run("FailedTransaction", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a test context
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		
		// Create a sample table for testing
		db.Exec("CREATE TABLE IF NOT EXISTS transaction_fail_test (id SERIAL PRIMARY KEY, value TEXT)")
		
		// Run a transaction that fails
		testErr := errors.New("test error")
		err := store.WithTransaction(c, func(tx *gorm.DB) error {
			// Insert a test record
			insertErr := tx.Exec("INSERT INTO transaction_fail_test (value) VALUES ('before_error')").Error
			if insertErr != nil {
				return insertErr
			}
			
			// Return an error to trigger rollback
			return testErr
		})
		
		// Verify transaction failed with the right error
		assert.Equal(t, testErr, err, "Expected test error from failed transaction")
		
		// Verify the record was rolled back
		var count int64
		db.Table("transaction_fail_test").Count(&count)
		assert.Equal(t, int64(0), count, "Expected no records after rollback")
	})
	
	t.Run("TransactionContext", func(t *testing.T) {
		tests.SetupTest(t, db)
		
		// Create a test context
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		
		// Test that we can get the tx from context
		err := store.WithTransaction(c, func(tx *gorm.DB) error {
			// Check that GetTxFromContext returns the transaction DB
			contextDB := store.GetTxFromContext(c)
			assert.NotNil(t, contextDB, "Expected DB from context to not be nil")
			return nil
		})
		
		assert.Nil(t, err, "Expected no error from transaction")
	})
}

package store_test

import (
	"errors"
	"meals/store"
	"meals/tests"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTransactions(t *testing.T) {
	db := tests.SetupTestSuite(t)

	t.Run("WithTransaction_Success", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test context
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/test", nil)

		// Test data
		testData := "test_data"
		var dataInTx string

		// Run a successful transaction with result
		result, err := store.TxWithResult(c, db, func(tx store.TxHandle) (interface{}, error) {
			// Get transaction from context
			txFromContext := store.GetTxFromContext(c)
			assert.NotNil(t, txFromContext, "Expected transaction to be in context")

			// Store test data in the transaction's context
			dataInTx = testData

			// Return success
			return testData, nil
		})

		// Verify transaction executed successfully
		assert.Nil(t, err, "Expected no error from successful transaction")
		assert.Equal(t, testData, result, "Expected result to match test data")
		assert.Equal(t, testData, dataInTx, "Expected data to be accessible in transaction")
	})

	t.Run("WithTransaction_Rollback", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test context
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/test", nil)

		// Create a sample table for testing transactions
		db.Exec("CREATE TABLE IF NOT EXISTS transaction_test (id SERIAL PRIMARY KEY, value TEXT)")

		// Test error
		testError := errors.New("test error")

		// Run a transaction that will fail
		result, err := store.TxWithResult(c, db, func(tx store.TxHandle) (interface{}, error) {
			// Insert a test record
			insertErr := tx.Exec("INSERT INTO transaction_test (value) VALUES ('test_value')").Error
			assert.Nil(t, insertErr, "Expected insert to succeed")

			// Count records in transaction
			var count int64
			tx.Table("transaction_test").Count(&count)
			assert.Equal(t, int64(1), count, "Expected one record in transaction")

			// Return error to trigger rollback
			return nil, testError
		})

		// Verify transaction was rolled back
		assert.Equal(t, testError, err, "Expected error from failed transaction")
		assert.Nil(t, result, "Expected nil result from failed transaction")

		// Verify the record was not committed
		var count int64
		db.Table("transaction_test").Count(&count)
		assert.Equal(t, int64(0), count, "Expected no records after rollback")
	})

	t.Run("GetTxFromContext_NoTransaction", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test context with Gin
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/test", nil)

		// Try to get transaction from context
		tx := store.GetTxFromContext(c)

		// Verify it returns the global DB instance when no transaction is in context
		assert.NotNil(t, tx, "Expected to get global DB instance when no transaction is in context")
		assert.Equal(t, store.DB, tx, "Expected to get global DB instance when no transaction is in context")
	})

	t.Run("NestedTransactions", func(t *testing.T) {
		tests.SetupTest(t, db)

		// Create a test context
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/test", nil)

		// Create test data
		outerData := "outer"
		innerData := "inner"
		var resultOuter, resultInner string

		// Run nested transactions
		result, err := store.TxWithResult(c, db, func(tx store.TxHandle) (interface{}, error) {
			// Store outer data
			resultOuter = outerData

			// Run inner transaction
			innerResult, innerErr := store.TxWithResult(c, tx, func(innerTx store.TxHandle) (interface{}, error) {
				// Store inner data
				resultInner = innerData
				return innerData, nil
			})

			// Verify inner transaction succeeded
			assert.Nil(t, innerErr, "Expected no error from inner transaction")
			assert.Equal(t, innerData, innerResult, "Expected inner result to match inner data")

			return resultOuter + ":" + resultInner, nil
		})

		// Verify outer transaction succeeded
		assert.Nil(t, err, "Expected no error from outer transaction")
		assert.Equal(t, outerData+":"+innerData, result, "Expected combined result from nested transactions")
		assert.Equal(t, outerData, resultOuter, "Expected outer data to be stored")
		assert.Equal(t, innerData, resultInner, "Expected inner data to be stored")
	})
}

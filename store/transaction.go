package store

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TxHandle is an alias for *gorm.DB to be used in transactions
type TxHandle = *gorm.DB

// TxFn represents a function that uses a transaction
type TxFn func(tx TxHandle) error

// TxFnWithResult represents a function that uses a transaction and returns a result
type TxFnWithResult func(tx TxHandle) (interface{}, error)

// WithTransactionResult executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
// This version supports returning a result along with an error.
// It also supports nested transactions using savepoints.
func WithTransactionResult(c *gin.Context, db *gorm.DB, fn TxFnWithResult) (interface{}, error) {
	// Check if we're already in a transaction
	tx, inTransaction := c.Request.Context().Value("tx").(*gorm.DB)

	if !inTransaction {
		// Start a new transaction if we're not already in one
		tx = db.Begin()
		if tx.Error != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
		}

		// Set transaction in context so it can be accessed by other functions
		if c != nil {
			ctx := context.WithValue(c.Request.Context(), "tx", tx)
			c.Request = c.Request.WithContext(ctx)
		}

		// Handle panic by rolling back
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				panic(r) // Re-throw panic after rollback
			}
		}()

		// Execute the function within the transaction
		result, err := fn(tx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		// Commit the transaction if no error occurred
		if err := tx.Commit().Error; err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return result, nil
	}

	// We're already in a transaction, so we'll use a savepoint for nesting
	savepointName := fmt.Sprintf("savepoint_%d", time.Now().UnixNano())
	err := tx.Exec("SAVEPOINT " + savepointName).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create savepoint: %w", err)
	}

	// Execute the function within the existing transaction
	result, err := fn(tx)
	if err != nil {
		// Rollback to the savepoint on error
		if rbErr := tx.Exec("ROLLBACK TO SAVEPOINT " + savepointName).Error; rbErr != nil {
			return nil, fmt.Errorf("failed to rollback to savepoint: %v (original error: %w)", rbErr, err)
		}
		return nil, err
	}

	// Release the savepoint on success
	if err := tx.Exec("RELEASE SAVEPOINT " + savepointName).Error; err != nil {
		return nil, fmt.Errorf("failed to release savepoint: %w", err)
	}

	return result, nil
}

// WithTransactionNoResult is a compatibility wrapper for the original WithTransaction function
// that doesn't return a result.
func WithTransactionNoResult(c *gin.Context, db *gorm.DB, fn TxFn) error {
	_, err := WithTransactionResult(c, db, func(tx TxHandle) (interface{}, error) {
		return nil, fn(tx)
	})
	return err
}

// Legacy compatibility function to match existing application code
// This allows tests to be updated without breaking the application
func WithTransaction(c *gin.Context, fn TxFn) error {
	return WithTransactionNoResult(c, DB, fn)
}

// TxWithResult is another name for WithTransactionResult with result to avoid naming conflict
func TxWithResult(c *gin.Context, db *gorm.DB, fn TxFnWithResult) (interface{}, error) {
	return WithTransactionResult(c, db, fn)
}

// GetTxFromContext extracts the transaction from the context if it exists
// Otherwise returns the provided DB instance or global DB if nil
func GetTxFromContext(c *gin.Context) *gorm.DB {
	if c == nil {
		return DB
	}

	tx, exists := c.Request.Context().Value("tx").(*gorm.DB)
	if !exists {
		return DB
	}

	return tx
}

package store

import (
	"context"
	"fmt"

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
func WithTransactionResult(c *gin.Context, db *gorm.DB, fn TxFnWithResult) (interface{}, error) {
	tx := db.Begin()
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

	result, err := fn(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
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

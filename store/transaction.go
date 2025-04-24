package store

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TxFn represents a function that uses a transaction
type TxFn func(tx *gorm.DB) error

// WithTransaction executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func WithTransaction(c *gin.Context, fn TxFn) error {
	tx := DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
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

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTxFromContext extracts the transaction from the context if it exists
// Otherwise returns the global DB instance
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

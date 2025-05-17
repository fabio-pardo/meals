package store

import (
	"meals/store"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TxWrapper provides test-specific transaction functions to maintain compatibility
// with the existing test suite while adapting to the updated transaction functions in the main app.

// WithTransactionForTest wraps store.TxWithResult to provide a compatible interface for tests
func WithTransactionForTest(c *gin.Context, db *gorm.DB, fn func(tx *gorm.DB) (interface{}, error)) (interface{}, error) {
	return store.TxWithResult(c, db, fn)
}

// WithTransactionNoResult wraps store.WithTransaction for tests that don't need a result
func WithTransactionNoResult(c *gin.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return store.WithTransactionNoResult(c, db, fn)
}

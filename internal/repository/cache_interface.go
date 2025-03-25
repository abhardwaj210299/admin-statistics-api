package repository

import (
	"time"
)

// Cache interface for caching responses
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, expiration time.Duration)
	Delete(key string)
}
package repository

import (
	"sync"
	"time"
)

// MockCache is a mock implementation of the Cache interface for testing
type MockCache struct {
	items            map[string]interface{}
	mu               sync.RWMutex
	GetCalls         []string
	SetCalls         map[string]interface{}
	DeleteCalls      []string
	GetShouldFail    bool
	GetCustomResults map[string]interface{}
}

// NewMockCache creates a new MockCache
func NewMockCache() *MockCache {
	return &MockCache{
		items:            make(map[string]interface{}),
		SetCalls:         make(map[string]interface{}),
		GetCustomResults: make(map[string]interface{}),
	}
}

// Get retrieves a value from the cache
func (c *MockCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.GetCalls = append(c.GetCalls, key)

	if c.GetShouldFail {
		return nil, false
	}

	// Check if we have a custom result for this key
	if result, ok := c.GetCustomResults[key]; ok {
		return result, true
	}

	// Fall back to the items map
	value, found := c.items[key]
	return value, found
}

// Set adds a value to the cache
func (c *MockCache) Set(key string, value interface{}, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = value
	c.SetCalls[key] = value
}

// Delete removes a value from the cache
func (c *MockCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.DeleteCalls = append(c.DeleteCalls, key)
	delete(c.items, key)
}
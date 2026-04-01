package claude

import (
	"fmt"
	"sync"
	"time"

	"realmweaver/engine"
)

// WorkerPool manages concurrent Claude API calls with rate limiting.
type WorkerPool struct {
	client       *Client
	haikuClient  *Client
	maxWorkers   int
	sem          chan struct{}
	cache        *ResponseCache
	mu           sync.Mutex
	stats        PoolStats
	userLastCall sync.Map // map[string]time.Time — per-user rate limiting
}

// PoolStats tracks API usage for monitoring.
type PoolStats struct {
	TemplateCount int64
	HaikuCount    int64
	SonnetCount   int64
	CacheHits     int64
	QueueDrops    int64
}

// ResponseCache stores generated narrations to avoid repeat API calls.
type ResponseCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	maxSize int
}

type CacheEntry struct {
	Response  *engine.ClaudeResponse
	CreatedAt time.Time
}

// NewWorkerPool creates a pool with the given concurrency limit.
func NewWorkerPool(maxWorkers int) *WorkerPool {
	sonnetClient := NewClient() // Uses CLAUDE_MODEL env (default: sonnet)

	haikuClient := NewClient()
	haikuClient.model = "claude-haiku-4-5-20251001" // Override to haiku

	return &WorkerPool{
		client:      sonnetClient,
		haikuClient: haikuClient,
		maxWorkers:  maxWorkers,
		sem:         make(chan struct{}, maxWorkers),
		cache:       NewResponseCache(1000),
	}
}

// NewResponseCache creates a cache with max entries.
func NewResponseCache(maxSize int) *ResponseCache {
	return &ResponseCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
	}
}

// Get returns a cached response if available and not expired.
func (c *ResponseCache) Get(key string) *engine.ClaudeResponse {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil
	}
	// Expire after 1 hour — delete stale entry
	if time.Since(entry.CreatedAt) > time.Hour {
		delete(c.entries, key)
		return nil
	}
	return entry.Response
}

// Set stores a response in the cache.
func (c *ResponseCache) Set(key string, resp *engine.ClaudeResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple eviction: if at capacity, remove oldest
	if len(c.entries) >= c.maxSize {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range c.entries {
			if oldestKey == "" || v.CreatedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.CreatedAt
			}
		}
		delete(c.entries, oldestKey)
	}

	c.entries[key] = &CacheEntry{
		Response:  resp,
		CreatedAt: time.Now(),
	}
}

// Generate routes a narration request to the appropriate tier.
// Returns the narration text. userID is used for per-user rate limiting.
func (wp *WorkerPool) Generate(
	userID string,
	tier NarrationTier,
	systemPrompt string,
	userMessage string,
	cacheKey string,
	result *engine.ActionResult,
	character *engine.Character,
	language string,
) (*engine.ClaudeResponse, error) {

	// Per-user rate limit: minimum 2 seconds between Claude calls
	if userID != "" && tier != TierTemplate {
		if last, ok := wp.userLastCall.Load(userID); ok {
			if time.Since(last.(time.Time)) < 2*time.Second {
				wp.mu.Lock()
				wp.stats.QueueDrops++
				wp.mu.Unlock()
				if result != nil && character != nil {
					text := TemplateNarrate(result, character, language)
					return &engine.ClaudeResponse{Narrative: text}, nil
				}
				return &engine.ClaudeResponse{
					Narrative: "The world continues around you...",
				}, nil
			}
		}
		wp.userLastCall.Store(userID, time.Now())
	}

	// Tier 0: Template — no API call
	if tier == TierTemplate && result != nil && character != nil {
		wp.mu.Lock()
		wp.stats.TemplateCount++
		wp.mu.Unlock()

		text := TemplateNarrate(result, character, language)
		return &engine.ClaudeResponse{Narrative: text}, nil
	}

	// Check cache
	if cacheKey != "" {
		if cached := wp.cache.Get(cacheKey); cached != nil {
			wp.mu.Lock()
			wp.stats.CacheHits++
			wp.mu.Unlock()
			return cached, nil
		}
	}

	// Try to acquire a worker slot (non-blocking with 5s timeout)
	select {
	case wp.sem <- struct{}{}:
		defer func() { <-wp.sem }()
	case <-time.After(5 * time.Second):
		// Queue full — return template fallback
		wp.mu.Lock()
		wp.stats.QueueDrops++
		wp.mu.Unlock()

		if result != nil && character != nil {
			text := TemplateNarrate(result, character, language)
			return &engine.ClaudeResponse{Narrative: text}, nil
		}
		return &engine.ClaudeResponse{
			Narrative: "The world continues around you...",
		}, nil
	}

	// Select client based on tier
	var client *Client
	switch tier {
	case TierHaiku:
		client = wp.haikuClient
		wp.mu.Lock()
		wp.stats.HaikuCount++
		wp.mu.Unlock()
	case TierSonnet:
		client = wp.client
		wp.mu.Lock()
		wp.stats.SonnetCount++
		wp.mu.Unlock()
	default:
		client = wp.haikuClient
	}

	resp, err := client.Generate(systemPrompt, userMessage)
	if err != nil {
		return nil, err
	}

	// Cache the response
	if cacheKey != "" {
		wp.cache.Set(cacheKey, resp)
	}

	return resp, nil
}

// GetStats returns current pool statistics.
func (wp *WorkerPool) GetStats() PoolStats {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.stats
}

// StatsString returns a human-readable stats summary.
func (wp *WorkerPool) StatsString() string {
	s := wp.GetStats()
	return fmt.Sprintf("Templates:%d Haiku:%d Sonnet:%d Cache:%d Drops:%d",
		s.TemplateCount, s.HaikuCount, s.SonnetCount, s.CacheHits, s.QueueDrops)
}

package weather

import (
	"context"
	"sync"
	"time"

	"github.com/shuv1824/recommender/internal/types"
)

// CachedWeatherService wraps WeatherService with caching
type CachedWeatherService struct {
	service     *WeatherService
	cache       []types.DistrictWeather
	lastUpdated time.Time
	cacheTTL    time.Duration
	mu          sync.RWMutex
	updating    bool
}

// NewCachedWeatherService creates a cached weather service
func NewCachedWeatherService(districts []types.District, cacheTTL time.Duration) *CachedWeatherService {
	return &CachedWeatherService{
		service:  NewWeatherService(districts),
		cacheTTL: cacheTTL,
	}
}

// GetTopCoolestAndCleanest returns cached data or fetches fresh data
func (c *CachedWeatherService) GetTopCoolestAndCleanest(ctx context.Context) ([]types.DistrictWeather, error) {
	c.mu.RLock()
	if c.cache != nil && time.Since(c.lastUpdated) < c.cacheTTL {
		result := make([]types.DistrictWeather, len(c.cache))
		copy(result, c.cache)
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	// Need to refresh cache
	c.mu.Lock()
	// Double-check after acquiring write lock
	if c.cache != nil && time.Since(c.lastUpdated) < c.cacheTTL {
		result := make([]types.DistrictWeather, len(c.cache))
		copy(result, c.cache)
		c.mu.Unlock()
		return result, nil
	}

	// Check if another goroutine is already updating
	if c.updating {
		// Return stale cache if available while update is in progress
		if c.cache != nil {
			result := make([]types.DistrictWeather, len(c.cache))
			copy(result, c.cache)
			c.mu.Unlock()
			return result, nil
		}
	}

	c.updating = true
	c.mu.Unlock()

	// Fetch fresh data
	data, err := c.service.GetTopCoolestAndCleanest(ctx)

	c.mu.Lock()
	c.updating = false
	if err == nil {
		c.cache = data
		c.lastUpdated = time.Now()
	}
	c.mu.Unlock()

	return data, err
}

// WarmCache pre-fetches data on startup
func (c *CachedWeatherService) WarmCache(ctx context.Context) error {
	_, err := c.GetTopCoolestAndCleanest(ctx)
	return err
}

// StartBackgroundRefresh starts a background goroutine to refresh cache periodically
func (c *CachedWeatherService) StartBackgroundRefresh(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(c.cacheTTL / 2) // Refresh before expiry
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Background refresh - don't block on errors
				refreshCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				c.GetTopCoolestAndCleanest(refreshCtx)
				cancel()
			}
		}
	}()
}

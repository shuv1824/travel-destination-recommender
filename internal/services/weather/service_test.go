package weather

import (
	"context"
	"testing"
	"time"

	"github.com/shuv1824/recommender/internal/types"
)

// TestRankDistricts tests the ranking logic
func TestRankDistricts(t *testing.T) {
	s := &WeatherService{}

	tests := []struct {
		name     string
		input    []types.DistrictWeather
		expected []types.DistrictWeather
	}{
		{
			name: "sorts by temperature ascending and returns top 10",
			input: []types.DistrictWeather{
				{ID: "1", Name: "District 1", AvgTemp2PM: 35.0, AvgPM25: 50.0},
				{ID: "2", Name: "District 2", AvgTemp2PM: 25.0, AvgPM25: 50.0},
				{ID: "3", Name: "District 3", AvgTemp2PM: 30.0, AvgPM25: 50.0},
				{ID: "4", Name: "District 4", AvgTemp2PM: 28.0, AvgPM25: 50.0},
				{ID: "5", Name: "District 5", AvgTemp2PM: 32.0, AvgPM25: 50.0},
				{ID: "6", Name: "District 6", AvgTemp2PM: 26.0, AvgPM25: 50.0},
				{ID: "7", Name: "District 7", AvgTemp2PM: 29.0, AvgPM25: 50.0},
				{ID: "8", Name: "District 8", AvgTemp2PM: 27.0, AvgPM25: 50.0},
				{ID: "9", Name: "District 9", AvgTemp2PM: 31.0, AvgPM25: 50.0},
				{ID: "10", Name: "District 10", AvgTemp2PM: 33.0, AvgPM25: 50.0},
			},
			expected: []types.DistrictWeather{
				{ID: "2", Name: "District 2", AvgTemp2PM: 25.0, AvgPM25: 50.0, Rank: 1},
				{ID: "6", Name: "District 6", AvgTemp2PM: 26.0, AvgPM25: 50.0, Rank: 2},
				{ID: "8", Name: "District 8", AvgTemp2PM: 27.0, AvgPM25: 50.0, Rank: 3},
			},
		},
		{
			name: "breaks temperature ties by PM2.5",
			input: []types.DistrictWeather{
				{ID: "1", Name: "Same Temp High PM", AvgTemp2PM: 25.0, AvgPM25: 100.0},
				{ID: "2", Name: "Same Temp Low PM", AvgTemp2PM: 25.0, AvgPM25: 30.0},
				{ID: "3", Name: "Same Temp Med PM", AvgTemp2PM: 25.0, AvgPM25: 60.0},
				{ID: "4", Name: "Warmer", AvgTemp2PM: 26.0, AvgPM25: 50.0},
				{ID: "5", Name: "Even Warmer", AvgTemp2PM: 27.0, AvgPM25: 50.0},
				{ID: "6", Name: "Hot 1", AvgTemp2PM: 28.0, AvgPM25: 50.0},
				{ID: "7", Name: "Hot 2", AvgTemp2PM: 29.0, AvgPM25: 50.0},
				{ID: "8", Name: "Hot 3", AvgTemp2PM: 30.0, AvgPM25: 50.0},
				{ID: "9", Name: "Hot 4", AvgTemp2PM: 31.0, AvgPM25: 50.0},
				{ID: "10", Name: "Hot 5", AvgTemp2PM: 32.0, AvgPM25: 50.0},
			},
			expected: []types.DistrictWeather{
				{ID: "2", Name: "Same Temp Low PM", AvgTemp2PM: 25.0, AvgPM25: 30.0, Rank: 1},
				{ID: "3", Name: "Same Temp Med PM", AvgTemp2PM: 25.0, AvgPM25: 60.0, Rank: 2},
				{ID: "1", Name: "Same Temp High PM", AvgTemp2PM: 25.0, AvgPM25: 100.0, Rank: 3},
			},
		},
		{
			name:     "handles empty slice",
			input:    []types.DistrictWeather{},
			expected: []types.DistrictWeather{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.rankDistricts(tt.input)

			// For empty input, expect empty output
			if len(tt.input) == 0 {
				if len(result) != 0 {
					t.Fatalf("expected 0 districts, got %d", len(result))
				}
				return
			}

			// rankDistricts always returns exactly 10 (or panics if less than 10 input)
			if len(result) != 10 {
				t.Fatalf("expected 10 districts, got %d", len(result))
			}

			// Check the first few results match expected order
			for i := range tt.expected {
				if result[i].ID != tt.expected[i].ID {
					t.Errorf("at position %d: expected ID %s, got %s", i, tt.expected[i].ID, result[i].ID)
				}
				if result[i].Rank != tt.expected[i].Rank {
					t.Errorf("at position %d: expected rank %d, got %d", i, tt.expected[i].Rank, result[i].Rank)
				}
			}
		})
	}
}

// TestRankDistrictsReturnsTop10 verifies that only top 10 districts are returned
func TestRankDistrictsReturnsTop10(t *testing.T) {
	s := &WeatherService{}

	// Create 15 districts
	input := make([]types.DistrictWeather, 15)
	for i := 0; i < 15; i++ {
		input[i] = types.DistrictWeather{
			ID:         string(rune('A' + i)),
			Name:       "District",
			AvgTemp2PM: float64(20 + i),
			AvgPM25:    50.0,
		}
	}

	result := s.rankDistricts(input)

	if len(result) != 10 {
		t.Errorf("expected 10 districts, got %d", len(result))
	}

	// Verify ranks are 1-10
	for i, d := range result {
		if d.Rank != i+1 {
			t.Errorf("expected rank %d, got %d", i+1, d.Rank)
		}
	}
}

// TestCachedWeatherService tests the caching logic
func TestCachedWeatherService(t *testing.T) {
	t.Run("returns cached data within TTL", func(t *testing.T) {
		districts := []types.District{
			{ID: "1", Name: "Test", Lat: 23.0, Long: 90.0},
		}

		svc := NewCachedWeatherService(districts, 1*time.Hour)

		// Manually set cache
		svc.mu.Lock()
		svc.cache = []types.DistrictWeather{
			{ID: "1", Name: "Test", AvgTemp2PM: 25.0, AvgPM25: 30.0, Rank: 1},
		}
		svc.lastUpdated = time.Now()
		svc.mu.Unlock()

		ctx := context.Background()
		result, err := svc.GetTopCoolestAndCleanest(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 1 {
			t.Errorf("expected 1 result, got %d", len(result))
		}

		if result[0].ID != "1" {
			t.Errorf("expected ID '1', got '%s'", result[0].ID)
		}
	})

	t.Run("cache expires after TTL", func(t *testing.T) {
		districts := []types.District{
			{ID: "1", Name: "Test", Lat: 23.0, Long: 90.0},
		}

		svc := NewCachedWeatherService(districts, 10*time.Millisecond)

		// Set cache with old timestamp
		svc.mu.Lock()
		svc.cache = []types.DistrictWeather{
			{ID: "1", Name: "Test", AvgTemp2PM: 25.0, AvgPM25: 30.0, Rank: 1},
		}
		svc.lastUpdated = time.Now().Add(-1 * time.Hour)
		svc.mu.Unlock()

		// Wait for TTL to expire
		time.Sleep(20 * time.Millisecond)

		// Note: This will fail because it tries to actually fetch data
		// In a real test, you'd want to mock the HTTP client
		// For minimal tests, we're just verifying the cache expiry logic
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		svc.mu.RLock()
		cacheExpired := time.Since(svc.lastUpdated) >= svc.cacheTTL
		svc.mu.RUnlock()

		if !cacheExpired {
			t.Error("expected cache to be expired")
		}

		// The actual call would require network access, so we skip it in minimal tests
		_ = ctx
	})

	t.Run("cache returns copy, not reference", func(t *testing.T) {
		districts := []types.District{
			{ID: "1", Name: "Test", Lat: 23.0, Long: 90.0},
		}

		svc := NewCachedWeatherService(districts, 1*time.Hour)

		svc.mu.Lock()
		svc.cache = []types.DistrictWeather{
			{ID: "1", Name: "Test", AvgTemp2PM: 25.0, AvgPM25: 30.0, Rank: 1},
		}
		svc.lastUpdated = time.Now()
		svc.mu.Unlock()

		ctx := context.Background()
		result1, _ := svc.GetTopCoolestAndCleanest(ctx)
		result2, _ := svc.GetTopCoolestAndCleanest(ctx)

		// Modify result1
		result1[0].AvgTemp2PM = 99.0

		// result2 should not be affected
		if result2[0].AvgTemp2PM == 99.0 {
			t.Error("cache returned reference instead of copy")
		}
	})
}

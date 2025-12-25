package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/shuv1824/recommender/internal/types"
)

type WeatherService struct {
	httpClient *http.Client
	districts  []types.District
}

func NewWeatherService(districts []types.District) *WeatherService {
	return &WeatherService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		districts: districts,
	}
}

// fetchResult holds the result of concurrent fetching
type fetchResult struct {
	District   types.District
	AvgTemp2PM float64
	AvgPM25    float64
	Err        error
}

// GetTopCoolestAndCleanest fetches weather data for all districts concurrently
// and returns the top 10 coolest and cleanest districts
func (s *WeatherService) GetTopCoolestAndCleanest(ctx context.Context) ([]types.DistrictWeather, error) {
	results := make(chan fetchResult, len(s.districts))
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrent requests (avoid rate limiting)
	semaphore := make(chan struct{}, 8) // Max 8 concurrent requests

	for _, district := range s.districts {
		wg.Add(1)
		go func(d types.District) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			avgTemp, avgPM25, err := s.fetchDistrictData(ctx, d)
			results <- fetchResult{
				District:   d,
				AvgTemp2PM: avgTemp,
				AvgPM25:    avgPM25,
				Err:        err,
			}
		}(district)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var districtWeathers []types.DistrictWeather
	for result := range results {
		if result.Err != nil {
			// Log error but continue with other districts
			fmt.Printf("Error fetching data for %s: %v\n", result.District.Name, result.Err)
			continue
		}

		districtWeathers = append(districtWeathers, types.DistrictWeather{
			ID:         result.District.ID,
			Name:       result.District.Name,
			BnName:     result.District.BnName,
			AvgTemp2PM: result.AvgTemp2PM,
			AvgPM25:    result.AvgPM25,
		})
	}

	//TODO: Calculate combined score and rank
	ranked := districtWeathers

	// Return top 10
	if len(ranked) > 10 {
		ranked = ranked[:10]
	}

	return ranked, nil
}

// fetchDistrictData fetches both weather and air quality data for a district
func (s *WeatherService) fetchDistrictData(ctx context.Context, d types.District) (float64, float64, error) {
	var (
		avgTemp float64
		avgPM25 float64
		tempErr error
		aqErr   error
		wg      sync.WaitGroup
	)

	// Fetch weather and air quality concurrently
	wg.Add(2)

	go func() {
		defer wg.Done()
		avgTemp, tempErr = s.fetchTemperature(ctx, d.Lat, d.Long)
	}()

	go func() {
		defer wg.Done()
		avgPM25, aqErr = s.fetchAirQuality(ctx, d.Lat, d.Long)
	}()

	wg.Wait()

	if tempErr != nil {
		return 0, 0, tempErr
	}
	if aqErr != nil {
		return 0, 0, aqErr
	}

	return avgTemp, avgPM25, nil
}

// fetchTemperature fetches 7-day hourly forecast and calculates avg temp at 2PM
func (s *WeatherService) fetchTemperature(ctx context.Context, lat, long float64) (float64, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&hourly=temperature_2m&timezone=auto",
		lat, long,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var data types.OpenMeteoForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	// Calculate average temperature at 2PM (14:00) for all 7 days
	var temps []float64
	for i, timeStr := range data.Hourly.Time {
		// Time format: "2025-12-25T14:00"
		if len(timeStr) >= 13 && timeStr[11:13] == "14" {
			if i < len(data.Hourly.Temperature2m) {
				temps = append(temps, data.Hourly.Temperature2m[i])
			}
		}
	}

	if len(temps) == 0 {
		return 0, fmt.Errorf("no 2PM temperature data found")
	}

	var sum float64
	for _, t := range temps {
		sum += t
	}
	return sum / float64(len(temps)), nil
}

// fetchAirQuality fetches air quality data and calculates avg PM2.5
func (s *WeatherService) fetchAirQuality(ctx context.Context, lat, long float64) (float64, error) {
	url := fmt.Sprintf(
		"https://air-quality-api.open-meteo.com/v1/air-quality?latitude=%.4f&longitude=%.4f&hourly=pm2_5&timezone=auto",
		lat, long,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("air quality API returned status %d", resp.StatusCode)
	}

	var data types.OpenMeteoAirQualityResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	// Calculate average PM2.5 at 2PM for all days
	var pm25Values []float64
	for i, timeStr := range data.Hourly.Time {
		if len(timeStr) >= 13 && timeStr[11:13] == "14" {
			if i < len(data.Hourly.PM25) {
				pm25Values = append(pm25Values, data.Hourly.PM25[i])
			}
		}
	}

	if len(pm25Values) == 0 {
		return 0, fmt.Errorf("no 2PM PM2.5 data found")
	}

	var sum float64
	for _, v := range pm25Values {
		sum += v
	}
	return sum / float64(len(pm25Values)), nil
}

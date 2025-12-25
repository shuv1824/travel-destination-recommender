package travel

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/shuv1824/recommender/internal/types"
)

type TravelService struct {
	httpClient *http.Client
	districts  map[string]types.District // Map by name for quick lookup
}

// NewTravelService creates a new travel service
func NewTravelService(districts []types.District) *TravelService {
	districtMap := make(map[string]types.District)
	for _, d := range districts {
		districtMap[d.Name] = d
	}

	return &TravelService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		districts: districtMap,
	}
}

// GetRecommendation compares current location with destination and returns recommendation
func (s *TravelService) GetRecommendation(ctx context.Context, req types.TravelRequest) (*types.TravelRecommendation, error) {
	// Validate travel date
	travelDate, err := time.Parse("2006-01-02", req.TravelDate)
	if err != nil {
		return nil, fmt.Errorf("invalid travel date format, use YYYY-MM-DD")
	}

	// Check if date is within forecast range (next 15 days)
	now := time.Now().Truncate(24 * time.Hour)
	maxDate := now.AddDate(0, 0, 15)
	if travelDate.Before(now) || travelDate.After(maxDate) {
		return nil, fmt.Errorf("travel date must be within the next 15 days")
	}

	// Get destination district
	destination, ok := s.districts[req.DestinationDistrictName]
	if !ok {
		return nil, fmt.Errorf("destination district not found")
	}

	// Fetch weather data for both locations concurrently
	type weatherResult struct {
		weather types.LocationWeather
		err     error
	}

	currentCh := make(chan weatherResult, 1)
	destCh := make(chan weatherResult, 1)

	// Get weather forecast for current location
	go func() {
		temp, pm25, err := s.fetchWeatherForDate(ctx, req.CurrentLocation.Lat, req.CurrentLocation.Long, req.TravelDate)
		name := req.CurrentLocation.Name
		if name == "" {
			name = "Current Location"
		}
		currentCh <- weatherResult{
			weather: types.LocationWeather{Name: name, Temp2PM: temp, PM25: pm25},
			err:     err,
		}
	}()

	// Get weather forecast for destination
	go func() {
		temp, pm25, err := s.fetchWeatherForDate(ctx, destination.Lat, destination.Long, req.TravelDate)
		destCh <- weatherResult{
			weather: types.LocationWeather{Name: destination.Name, Temp2PM: temp, PM25: pm25},
			err:     err,
		}
	}()

	currentResult := <-currentCh
	destResult := <-destCh

	if currentResult.err != nil {
		return nil, fmt.Errorf("failed to fetch current location weather: %w", currentResult.err)
	}
	if destResult.err != nil {
		return nil, fmt.Errorf("failed to fetch destination weather: %w", destResult.err)
	}

	// Calculate differences
	tempDiff := math.Round((currentResult.weather.Temp2PM-destResult.weather.Temp2PM)*100) / 100
	pm25Diff := math.Round((currentResult.weather.PM25-destResult.weather.PM25)*100) / 100

	// Determine recommendation
	isCooler := destResult.weather.Temp2PM < currentResult.weather.Temp2PM
	isCleaner := destResult.weather.PM25 < currentResult.weather.PM25

	recommended := "Not Recommended"
	if isCleaner && isCooler {
		recommended = "Recommended"
	}

	reason := "Reason" // TODO: generateReason

	return &types.TravelRecommendation{
		Recommendation:     recommended,
		Reason:             reason,
		TravelDate:         req.TravelDate,
		CurrentWeather:     currentResult.weather,
		DestinationWeather: destResult.weather,
		TempDifference:     tempDiff,
		PM25Difference:     pm25Diff,
	}, nil
}

// fetchWeatherForDate fetches temperature and PM2.5 at 2PM for a specific date
func (s *TravelService) fetchWeatherForDate(ctx context.Context, lat, long float64, date string) (float64, float64, error) {
	type result struct {
		value float64
		err   error
	}

	tempCh := make(chan result, 1)
	pm25Ch := make(chan result, 1)

	// Fetch temperature
	go func() {
		temp, err := s.fetchTemperature(ctx, lat, long, date)
		tempCh <- result{value: temp, err: err}
	}()

	// Fetch air quality
	go func() {
		pm25, err := s.fetchPM25(ctx, lat, long, date)
		pm25Ch <- result{value: pm25, err: err}
	}()

	tempResult := <-tempCh
	pm25Result := <-pm25Ch

	if tempResult.err != nil {
		return 0, 0, tempResult.err
	}
	if pm25Result.err != nil {
		return 0, 0, pm25Result.err
	}

	return tempResult.value, pm25Result.value, nil
}

// fetchTemperature fetches temperature at 2PM for a specific date
func (s *TravelService) fetchTemperature(ctx context.Context, lat, long float64, date string) (float64, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&hourly=temperature_2m&start_date=%s&end_date=%s&timezone=auto",
		lat, long, date, date,
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

	var data struct {
		Hourly struct {
			Time          []string  `json:"time"`
			Temperature2m []float64 `json:"temperature_2m"`
		} `json:"hourly"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	// Find temperature at 2PM (14:00)
	for i, timeStr := range data.Hourly.Time {
		if len(timeStr) >= 13 && timeStr[11:13] == "14" {
			if i < len(data.Hourly.Temperature2m) {
				return math.Round(data.Hourly.Temperature2m[i]*100) / 100, nil
			}
		}
	}

	return 0, fmt.Errorf("no 2PM temperature data found")
}

// fetchPM25 fetches PM2.5 at 2PM for a specific date
func (s *TravelService) fetchPM25(ctx context.Context, lat, long float64, date string) (float64, error) {
	url := fmt.Sprintf(
		"https://air-quality-api.open-meteo.com/v1/air-quality?latitude=%.4f&longitude=%.4f&hourly=pm2_5&start_date=%s&end_date=%s&timezone=auto",
		lat, long, date, date,
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

	var data struct {
		Hourly struct {
			Time []string  `json:"time"`
			PM25 []float64 `json:"pm2_5"`
		} `json:"hourly"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	// Find PM2.5 at 2PM (14:00)
	for i, timeStr := range data.Hourly.Time {
		if len(timeStr) >= 13 && timeStr[11:13] == "14" {
			if i < len(data.Hourly.PM25) {
				return math.Round(data.Hourly.PM25[i]*100) / 100, nil
			}
		}
	}

	return 0, fmt.Errorf("no 2PM PM2.5 data found")
}

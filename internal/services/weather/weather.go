package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

// fetchTemperature fetches 7-day hourly forecast and calculates avg temp at 2PM
func (s *WeatherService) FetchTemperature(ctx context.Context, lat, long float64) (float64, error) {
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
		// Time format: "2024-01-15T14:00"
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
func (s *WeatherService) FetchAirQuality(ctx context.Context, lat, long float64) (float64, error) {
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

package travel

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shuv1824/recommender/internal/types"
)

// mockTransport is a mock HTTP transport for testing
type mockTransport struct {
	responses map[string]string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()

	// Determine which mock response to return based on URL
	var body string
	if strings.Contains(url, "api.open-meteo.com/v1/forecast") {
		// Temperature API
		if strings.Contains(url, "latitude=23.8103") {
			// Current location (Dhaka)
			body = m.responses["temp_current"]
		} else if strings.Contains(url, "latitude=22.3569") {
			// Cox's Bazar
			body = m.responses["temp_dest"]
		}
	} else if strings.Contains(url, "air-quality-api.open-meteo.com") {
		// Air quality API
		if strings.Contains(url, "latitude=23.8103") {
			// Current location (Dhaka)
			body = m.responses["pm25_current"]
		} else if strings.Contains(url, "latitude=22.3569") {
			// Cox's Bazar
			body = m.responses["pm25_dest"]
		}
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}, nil
}

func TestGetRecommendation(t *testing.T) {
	// Create tomorrow's date for testing
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	tests := []struct {
		name               string
		request            types.TravelRequest
		mockResponses      map[string]string
		expectedRecommend  string
		expectError        bool
		errorContains      string
	}{
		{
			name: "recommended when destination is cooler and cleaner",
			request: types.TravelRequest{
				CurrentLocation: types.Location{
					Lat:  23.8103,
					Long: 90.4125,
					Name: "Dhaka",
				},
				DestinationDistrictName: "Cox's Bazar",
				TravelDate:              tomorrow,
			},
			mockResponses: map[string]string{
				"temp_current": `{"hourly":{"time":["` + tomorrow + `T14:00"],"temperature_2m":[35.5]}}`,
				"temp_dest":    `{"hourly":{"time":["` + tomorrow + `T14:00"],"temperature_2m":[28.0]}}`,
				"pm25_current": `{"hourly":{"time":["` + tomorrow + `T14:00"],"pm2_5":[75.0]}}`,
				"pm25_dest":    `{"hourly":{"time":["` + tomorrow + `T14:00"],"pm2_5":[25.0]}}`,
			},
			expectedRecommend: "Recommended",
			expectError:       false,
		},
		{
			name: "not recommended when destination is hotter",
			request: types.TravelRequest{
				CurrentLocation: types.Location{
					Lat:  23.8103,
					Long: 90.4125,
					Name: "Dhaka",
				},
				DestinationDistrictName: "Cox's Bazar",
				TravelDate:              tomorrow,
			},
			mockResponses: map[string]string{
				"temp_current": `{"hourly":{"time":["` + tomorrow + `T14:00"],"temperature_2m":[28.0]}}`,
				"temp_dest":    `{"hourly":{"time":["` + tomorrow + `T14:00"],"temperature_2m":[35.0]}}`,
				"pm25_current": `{"hourly":{"time":["` + tomorrow + `T14:00"],"pm2_5":[75.0]}}`,
				"pm25_dest":    `{"hourly":{"time":["` + tomorrow + `T14:00"],"pm2_5":[25.0]}}`,
			},
			expectedRecommend: "Not Recommended",
			expectError:       false,
		},
		{
			name: "invalid date format returns error",
			request: types.TravelRequest{
				CurrentLocation: types.Location{
					Lat:  23.8103,
					Long: 90.4125,
				},
				DestinationDistrictName: "Cox's Bazar",
				TravelDate:              "invalid-date",
			},
			expectError:   true,
			errorContains: "invalid travel date format",
		},
		{
			name: "date in past returns error",
			request: types.TravelRequest{
				CurrentLocation: types.Location{
					Lat:  23.8103,
					Long: 90.4125,
				},
				DestinationDistrictName: "Cox's Bazar",
				TravelDate:              "2020-01-01",
			},
			expectError:   true,
			errorContains: "travel date must be within the next 7 days",
		},
		{
			name: "date too far in future returns error",
			request: types.TravelRequest{
				CurrentLocation: types.Location{
					Lat:  23.8103,
					Long: 90.4125,
				},
				DestinationDistrictName: "Cox's Bazar",
				TravelDate:              time.Now().AddDate(0, 0, 10).Format("2006-01-02"),
			},
			expectError:   true,
			errorContains: "travel date must be within the next 7 days",
		},
		{
			name: "invalid district returns error",
			request: types.TravelRequest{
				CurrentLocation: types.Location{
					Lat:  23.8103,
					Long: 90.4125,
				},
				DestinationDistrictName: "NonExistent District",
				TravelDate:              tomorrow,
			},
			expectError:   true,
			errorContains: "destination district not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service with mock HTTP client
			districts := []types.District{
				{
					ID:   "1",
					Name: "Cox's Bazar",
					Lat:  22.3569,
					Long: 91.7832,
				},
			}

			service := NewTravelService(districts)

			// Replace HTTP client with mock only for success cases
			if !tt.expectError || tt.errorContains == "destination district not found" {
				service.httpClient = &http.Client{
					Transport: &mockTransport{responses: tt.mockResponses},
				}
			}

			// Call the method
			ctx := context.Background()
			result, err := service.GetRecommendation(ctx, tt.request)

			// Check error cases
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing '%s', got nil", tt.errorContains)
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
				return
			}

			// Check success cases
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Recommendation != tt.expectedRecommend {
				t.Errorf("expected recommendation '%s', got '%s'", tt.expectedRecommend, result.Recommendation)
			}

			// Verify basic fields are populated
			if result.TravelDate != tt.request.TravelDate {
				t.Errorf("expected travel date '%s', got '%s'", tt.request.TravelDate, result.TravelDate)
			}

			if result.Reason == "" {
				t.Error("expected non-empty reason")
			}

			if result.CurrentWeather.Name == "" {
				t.Error("expected non-empty current weather name")
			}

			if result.DestinationWeather.Name == "" {
				t.Error("expected non-empty destination weather name")
			}
		})
	}
}

func TestGenerateReason(t *testing.T) {
	s := &TravelService{}

	tests := []struct {
		name       string
		isCooler   bool
		isCleaner  bool
		tempDiff   float64
		pm25Diff   float64
		destName   string
		shouldContain []string
	}{
		{
			name:      "cooler and cleaner with significant differences",
			isCooler:  true,
			isCleaner: true,
			tempDiff:  5.5,
			pm25Diff:  20.0,
			destName:  "Cox's Bazar",
			shouldContain: []string{"Cox's Bazar", "cooler", "better air quality"},
		},
		{
			name:      "hotter and worse air quality",
			isCooler:  false,
			isCleaner: false,
			tempDiff:  -4.0,
			pm25Diff:  -18.0,
			destName:  "Dhaka",
			shouldContain: []string{"Dhaka", "hotter", "worse air quality"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := s.generateReason(tt.isCooler, tt.isCleaner, tt.tempDiff, tt.pm25Diff, tt.destName)

			for _, substr := range tt.shouldContain {
				if !strings.Contains(reason, substr) {
					t.Errorf("expected reason to contain '%s', got: %s", substr, reason)
				}
			}
		})
	}
}

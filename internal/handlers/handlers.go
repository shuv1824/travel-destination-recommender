package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/shuv1824/recommender/internal/response"
	"github.com/shuv1824/recommender/internal/services/weather"
	"github.com/shuv1824/recommender/internal/types"
)

type WeatherHandler struct {
	service *weather.WeatherService
}

func NewWeatherHandler(service *weather.WeatherService) *WeatherHandler {
	return &WeatherHandler{service: service}
}

type TopDestinationsResponse struct {
	GeneratedAt  string                  `json:"generated_at"`
	Description  string                  `json:"description"`
	Destinations []types.DistrictWeather `json:"destinations"`
}

// Health returns a simple health check response
func Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// GetTopDestinations returns top 10 coolest and cleanest districts
func (h *WeatherHandler) GetTopDestinations(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 9000*time.Millisecond)
	defer cancel()

	start := time.Now()

	destinations, err := h.service.GetTopCoolestAndCleanest(ctx)
	if err != nil {
		// If context deadline exceeded, return cached or error
		if ctx.Err() == context.DeadlineExceeded {
			response.ErrorJSON(w, http.StatusGatewayTimeout, "request timeout - try again")
			return
		}
		response.ErrorJSON(w, http.StatusInternalServerError, "failed to fetch weather data")
		return
	}

	resp := TopDestinationsResponse{
		GeneratedAt:  time.Now().Format(time.RFC3339),
		Description:  "Top 10 coolest and cleanest districts in Bangladesh based on 7-day forecast (2PM temperature and PM2.5 levels)",
		Destinations: destinations,
	}

	// Add response time header for debugging
	w.Header().Set("X-Response-Time", time.Since(start).String())

	response.JSON(w, http.StatusOK, resp)
}

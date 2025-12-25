package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/shuv1824/recommender/internal/response"
	"github.com/shuv1824/recommender/internal/services/travel"
	"github.com/shuv1824/recommender/internal/services/weather"
	"github.com/shuv1824/recommender/internal/types"
)

type RecommendationHandler struct {
	weatherService *weather.CachedWeatherService
	travelService  *travel.TravelService
}

func NewRecommendationHandler(weatherService *weather.CachedWeatherService, travelService *travel.TravelService) *RecommendationHandler {
	return &RecommendationHandler{
		weatherService: weatherService,
		travelService:  travelService,
	}
}

// Health returns a simple health check response
func Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// GetTopDestinations returns top 10 coolest and cleanest districts
func (h *RecommendationHandler) GetTopDestinations(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 490*time.Millisecond)
	defer cancel()

	start := time.Now()

	destinations, err := h.weatherService.GetTopCoolestAndCleanest(ctx)
	if err != nil {
		// If context deadline exceeded, return cached or error
		if ctx.Err() == context.DeadlineExceeded {
			response.ErrorJSON(w, http.StatusGatewayTimeout, "request timeout - try again")
			return
		}
		response.ErrorJSON(w, http.StatusInternalServerError, "failed to fetch weather data")
		return
	}

	resp := types.TopDestinationsResponse{
		GeneratedAt:  time.Now().Format(time.RFC3339),
		Description:  "Top 10 coolest and cleanest districts in Bangladesh based on 7-day forecast (2PM temperature and PM2.5 levels)",
		Destinations: destinations,
	}

	// Add response time header for debugging
	w.Header().Set("X-Response-Time", time.Since(start).String())

	response.JSON(w, http.StatusOK, resp)
}

func (h *RecommendationHandler) GetRecommendation(w http.ResponseWriter, r *http.Request) {
	var body types.TravelRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if body.CurrentLocation.Lat == 0 && body.CurrentLocation.Long == 0 {
		response.ErrorJSON(w, http.StatusBadRequest, "current_location lat and long are required")
		return
	}
	if body.DestinationDistrictName == "" {
		response.ErrorJSON(w, http.StatusBadRequest, "destination_district_id is required")
		return
	}
	if body.TravelDate == "" {
		response.ErrorJSON(w, http.StatusBadRequest, "travel_date is required (format: YYYY-MM-DD)")
		return
	}

	// Build request
	req := types.TravelRequest{
		CurrentLocation: types.Location{
			Lat:  body.CurrentLocation.Lat,
			Long: body.CurrentLocation.Long,
			Name: body.CurrentLocation.Name,
		},
		DestinationDistrictName: body.DestinationDistrictName,
		TravelDate:              body.TravelDate,
	}

	start := time.Now()

	recommendation, err := h.travelService.GetRecommendation(r.Context(), req)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Add response time header
	w.Header().Set("X-Response-Time", time.Since(start).String())

	response.JSON(w, http.StatusOK, recommendation)
}

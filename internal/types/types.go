package types

type RawDistrict struct {
	ID         string `json:"id"`
	DivisionID string `json:"division_id"`
	Name       string `json:"name"`
	BnName     string `json:"bn_name"`
	Lat        string `json:"lat"`
	Long       string `json:"long"`
}

type District struct {
	ID         string  `json:"id"`
	DivisionID string  `json:"division_id"`
	Name       string  `json:"name"`
	BnName     string  `json:"bn_name"`
	Lat        float64 `json:"lat"`
	Long       float64 `json:"long"`
}

type GeoData struct {
	Districts []RawDistrict `json:"districts"`
}

type DistrictWeather struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	AvgTemp2PM float64 `json:"avg_temp_2pm_celsius"`
	AvgPM25    float64 `json:"avg_pm25"`
	Rank       int     `json:"rank"`
}

type Location struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
	Name string  `json:"name,omitempty"`
}

type TopDestinationsResponse struct {
	GeneratedAt  string            `json:"generated_at"`
	Description  string            `json:"description"`
	Destinations []DistrictWeather `json:"destinations"`
}

type LocationWeather struct {
	Name    string  `json:"name"`
	Temp2PM float64 `json:"temp_2pm_celsius"`
	PM25    float64 `json:"pm25"`
}

type TravelRequest struct {
	CurrentLocation         Location `json:"current_location"`
	DestinationDistrictName string   `json:"destination_district"`
	TravelDate              string   `json:"travel_date"` // Format: YYYY-MM-DD
}

// TravelRequestBody is the request body for travel recommendation
type TravelRequestBody struct {
	CurrentLocation struct {
		Lat  float64 `json:"lat"`
		Long float64 `json:"long"`
		Name string  `json:"name,omitempty"`
	} `json:"current_location"`
	DestinationDistrictName string `json:"destination_district"`
	TravelDate              string `json:"travel_date"`
}

// TravelRecommendation is the API response
type TravelRecommendation struct {
	Recommendation     string          `json:"recommendation"`
	Reason             string          `json:"reason"`
	TravelDate         string          `json:"travel_date"`
	CurrentWeather     LocationWeather `json:"current_location"`
	DestinationWeather LocationWeather `json:"destination"`
	TempDifference     float64         `json:"temp_difference_celsius"`
	PM25Difference     float64         `json:"pm25_difference"`
}

// OpenMeteoForecastResponse represents the weather API response
type OpenMeteoForecastResponse struct {
	Hourly struct {
		Time          []string  `json:"time"`
		Temperature2m []float64 `json:"temperature_2m"`
	} `json:"hourly"`
}

// OpenMeteoAirQualityResponse represents the air quality API response
type OpenMeteoAirQualityResponse struct {
	Hourly struct {
		Time []string  `json:"time"`
		PM25 []float64 `json:"pm2_5"`
	} `json:"hourly"`
}

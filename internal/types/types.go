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
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	BnName         string  `json:"bn_name"`
	AvgTemp2PM     float64 `json:"avg_temp_2pm_celsius"`
	AvgPM25        float64 `json:"avg_pm25"`
	CoolnessRank   int     `json:"coolness_rank,omitempty"`
	AirQualityRank int     `json:"air_quality_rank,omitempty"`
	CombinedScore  float64 `json:"combined_score"` // Lower is better
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

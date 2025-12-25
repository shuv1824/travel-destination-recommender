package geodata

import (
	"encoding/json"
	"os"
	"sync"
)

type District struct {
	ID         string `json:"id"`
	DivisionID string `json:"division_id"`
	Name       string `json:"name"`
	BnName     string `json:"bn_name"`
	Lat        string `json:"lat"`
	Long       string `json:"long"`
}

type GeoData struct {
	Districts []District `json:"districts"`
}

var (
	data     GeoData
	loadOnce sync.Once
	loadErr  error
)

// Load reads the JSON file once. Safe to call multiple times.
func Load(filepath string) error {
	loadOnce.Do(func() {
		file, err := os.Open(filepath)
		if err != nil {
			loadErr = err
			return
		}
		defer file.Close()

		loadErr = json.NewDecoder(file).Decode(&data)
	})
	return loadErr
}

func Districts() []District {
	return data.Districts
}

func FindDistrict(id string) *District {
	for i := range data.Districts {
		if data.Districts[i].ID == id {
			return &data.Districts[i]
		}
	}
	return nil
}

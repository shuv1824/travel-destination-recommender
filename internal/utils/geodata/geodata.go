package geodata

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/shuv1824/recommender/internal/types"
)

var (
	data     []types.District
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

		var raw types.GeoData
		if err := json.NewDecoder(file).Decode(&raw); err != nil {
			loadErr = err
			return
		}

		// Convert to weather.District with parsed coordinates
		data = make([]types.District, 0, len(raw.Districts))
		for _, d := range raw.Districts {
			lat, err := strconv.ParseFloat(d.Lat, 64)
			if err != nil {
				continue
			}
			long, err := strconv.ParseFloat(d.Long, 64)
			if err != nil {
				continue
			}

			data = append(data, types.District{
				ID:         d.ID,
				DivisionID: d.DivisionID,
				Name:       d.Name,
				BnName:     d.BnName,
				Lat:        lat,
				Long:       long,
			})
		}
	})

	return loadErr
}

func Districts() []types.District {
	return data
}

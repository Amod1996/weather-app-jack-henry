package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

const (
	apiURL           = "https://api.weather.gov/points"
	cacheDuration    = 30 * time.Minute // Adjust cache duration as needed
	evictionInterval = 1 * time.Minute  // Interval at which cache is checked for eviction
)

// ForecastResponse represents the structure of the weather data returned by the National Weather Service API
type ForecastResponse struct {
	Properties struct {
		Periods []struct {
			Name            string `json:"name"`
			Temperature     int    `json:"temperature"`
			TemperatureUnit string `json:"temperatureUnit"`
			ShortForecast   string `json:"shortForecast"`
		} `json:"periods"`
	} `json:"properties"`
}

// CacheEntry represents a cache entry with weather data and a timestamp
type CacheEntry struct {
	data      ForecastResponse
	timestamp time.Time
}

var (
	cache = make(map[string]CacheEntry)
	mu    sync.RWMutex
)

// isValidLatitude checks if the given latitude is valid
func isValidLatitude(lat float64) bool {
	return lat >= -90 && lat <= 90
}

// isValidLongitude checks if the given longitude is valid
func isValidLongitude(lon float64) bool {
	return lon >= -180 && lon <= 180
}

// getWeather handles the HTTP request to fetch weather data
func getWeather(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	if latStr == "" || lonStr == "" {
		http.Error(w, "Please provide latitude and longitude parameters", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || !isValidLatitude(lat) {
		http.Error(w, "Invalid latitude value", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil || !isValidLongitude(lon) {
		http.Error(w, "Invalid longitude value", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("%s:%s", latStr, lonStr)

	// Check cache
	mu.RLock()
	entry, exists := cache[key]
	mu.RUnlock()

	if exists && time.Since(entry.timestamp) < cacheDuration {
		fmt.Println("cache exists")
		respondWithWeather(w, entry.data)
		return
	}

	// Fetch data from National Weather Service API
	forecastURL, err := getForecastURL(lat, lon)
	if err != nil {
		http.Error(w, "Failed to get forecast URL", http.StatusInternalServerError)
		return
	}

	resp, err := http.Get(forecastURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to get weather data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var forecastResponse ForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&forecastResponse); err != nil {
		http.Error(w, "Failed to parse weather data", http.StatusInternalServerError)
		return
	}

	// Cache the response
	mu.Lock()
	cache[key] = CacheEntry{
		data:      forecastResponse,
		timestamp: time.Now(),
	}
	mu.Unlock()

	respondWithWeather(w, forecastResponse)
}

// getForecastURL retrieves the forecast URL for the provided latitude and longitude
func getForecastURL(lat, lon float64) (string, error) {
	url := fmt.Sprintf("%s/%f,%f", apiURL, lat, lon)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get forecast URL")
	}
	defer resp.Body.Close()

	var result struct {
		Properties struct {
			Forecast string `json:"forecast"`
		} `json:"properties"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse forecast URL")
	}

	return result.Properties.Forecast, nil
}

// respondWithWeather sends the weather data as a JSON response
func respondWithWeather(w http.ResponseWriter, forecastResponse ForecastResponse) {
	todayForecast := forecastResponse.Properties.Periods[0]

	temperatureDescription := characterizeTemperature(todayForecast.Temperature)

	response := map[string]interface{}{
		"shortForecast":      todayForecast.ShortForecast,
		"temperature":        todayForecast.Temperature,
		"temperatureUnit":    todayForecast.TemperatureUnit,
		"weatherDescription": temperatureDescription,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// characterizeTemperature classifies the temperature into "hot", "cold", or "moderate"
func characterizeTemperature(temp int) string {
	switch {
	case temp <= 40:
		return "cold"
	case temp > 40 && temp <= 55:
		return "chilly"
	case temp > 55 && temp <= 75:
		return "moderate"
	case temp > 75 && temp <= 90:
		return "hot"
	case temp > 90:
		return "very hot"
	default:
		return "unknown"
	}
}

// startCacheEviction starts a goroutine to periodically evict expired cache entries
func startCacheEviction() {
	for {
		time.Sleep(evictionInterval)
		mu.Lock()
		for key, entry := range cache {
			if time.Since(entry.timestamp) > cacheDuration {
				delete(cache, key)
				fmt.Println("Cache delete")
			}
		}
		mu.Unlock()
	}
}

// main is the entry point of the application
func main() {
	go startCacheEviction()
	r := mux.NewRouter()
	r.HandleFunc("/weather", getWeather).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

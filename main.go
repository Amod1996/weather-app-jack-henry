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
	apiURL           = "https://api.openweathermap.org/data/2.5/weather"
	apiKey           = "tempKey"        // Replace with your actual API key
	cacheDuration    = 30 * time.Minute // Adjust cache duration as needed
	evictionInterval = 1 * time.Minute  // Interval at which cache is checked for eviction
)

// WeatherResponse represents the structure of the weather data returned by the API
type WeatherResponse struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
		SeaLevel  int     `json:"sea_level"`
		GrndLevel int     `json:"grnd_level"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int64 `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

// CacheEntry represents a cache entry with weather data and a timestamp
type CacheEntry struct {
	data      WeatherResponse
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

	// Fetch data from OpenWeather API
	url := fmt.Sprintf("%s?lat=%f&lon=%f&appid=%s&units=imperial", apiURL, lat, lon, apiKey)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to get weather data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var weatherResponse WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
		http.Error(w, "Failed to parse weather data", http.StatusInternalServerError)
		return
	}

	// Cache the response
	mu.Lock()
	cache[key] = CacheEntry{
		data:      weatherResponse,
		timestamp: time.Now(),
	}
	mu.Unlock()

	respondWithWeather(w, weatherResponse)
}

// respondWithWeather sends the weather data as a JSON response
func respondWithWeather(w http.ResponseWriter, weatherResponse WeatherResponse) {
	condition := weatherResponse.Weather[0].Main
	temp := weatherResponse.Main.Temp

	var temperatureDescription string
	switch {
	case temp <= 32:
		temperatureDescription = "very cold"
	case temp > 32 && temp <= 50:
		temperatureDescription = "cold"
	case temp > 50 && temp <= 65:
		temperatureDescription = "chilly"
	case temp > 65 && temp <= 85:
		temperatureDescription = "warm"
	case temp > 85 && temp <= 100:
		temperatureDescription = "hot"
	default:
		temperatureDescription = "extreme heat"
	}

	response := map[string]interface{}{
		"location": map[string]interface{}{
			"name":      weatherResponse.Name,
			"latitude":  weatherResponse.Coord.Lat,
			"longitude": weatherResponse.Coord.Lon,
		},
		"current": map[string]interface{}{
			"condition":              condition,
			"temperature":            fmt.Sprintf("%.2f", temp),
			"temperatureDescription": temperatureDescription,
			"humidity":               weatherResponse.Main.Humidity,
			"pressure":               weatherResponse.Main.Pressure,
			"windSpeed":              weatherResponse.Wind.Speed,
			"windDirection":          weatherResponse.Wind.Deg,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

// getCache handles the HTTP request to return the current cache state
func getCache(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	response := make(map[string]interface{})
	for key, entry := range cache {
		response[key] = map[string]interface{}{
			"data":      entry.data,
			"timestamp": entry.timestamp,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// main is the entry point of the application
func main() {
	go startCacheEviction()
	r := mux.NewRouter()
	r.HandleFunc("/weather", getWeather).Methods("GET")
	r.HandleFunc("/cache", getCache).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

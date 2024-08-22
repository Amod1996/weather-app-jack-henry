# Weather API Service with Caching

This Go application provides a simple weather API service that retrieves weather data from the OpenWeather API, caches the results, and serves them to clients. It uses the `gorilla/mux` package for routing and supports cache eviction to ensure that cached data is periodically refreshed.

## Features

- Fetches weather data from the OpenWeather API.
- Caches the fetched weather data for a specified duration.
- Periodically evicts stale cache entries.
- Provides weather data based on latitude and longitude.
- Includes an endpoint to view the current cache state.

## Prerequisites

- Go 1.16 or later
- An API key from OpenWeather (replace the placeholder `apiKey` in the code with your actual API key)

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/Amod1996/weather-app-jack-henry.git
    cd weather-app-jack-henry
    ```
2. Set the `apiKey` variable in the main.go to your OpenWeather API key.   

3. Install the dependencies:
    ```sh
    go mod tidy
    ```

4. Build the application:
    ```sh
    go build
    ```

## Usage


1. Run the application(without docker):
    ```sh
    ./weather-app-jack-henry
    ```

2. The server will start on port 8080 by default. You can set a different port by setting the `PORT` environment variable:
    ```sh
    export PORT=8080
    ./weather-app-jack-henry
    ```
3. You can also run this service in Docker using :
    ```sh
    docker-compose up
    ``` 

## Endpoints

### Get Weather Data

- **URL:** `http://localhost:8080/weather?lat=<lat>&long=<long>`
- **Method:** `GET`
- **Query Parameters:**
    - `lat` (float): Latitude of the location.
    - `lon` (float): Longitude of the location.
- **Response:**
  ```json
  {
    "current": {
      "condition": "Clear",
      "humidity": 73,
      "pressure": 1016,
      "temperature": "52.59",
      "temperatureDescription": "chilly",
      "windDirection": 350,
      "windSpeed": 6.91
    },
    "location": {
      "name": "City Name",
      "latitude": 12.34,
      "longitude": 56.78
    }
  }
  
### Get Active Cache
- **URL:** `http://localhost:8080/cache`


# Weather API Service with Caching

This Go application provides a simple weather API service that retrieves weather data from the National Weather Service API, caches the results, and serves them to clients. It uses the `gorilla/mux` package for routing and supports cache eviction to ensure that cached data is periodically refreshed.

## Features

- Fetches weather data from the National Weather Service API.
- Caches the fetched weather data for a specified duration.
- Periodically evicts stale cache entries.
- Provides weather data based on latitude and longitude.
- Categorizes temperature as "cold","chilly", "moderate", "hot" or "very hot".
- Returns a short forecast for the specified location.
- Provides an endpoint to view the current cache state.

## Prerequisites

- Go 1.16 or later
- Docker should be running and installed if you are running this app in docker

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/Amod1996/weather-app-jack-henry.git
    cd weather-app-jack-henry
    ```

2. Install the dependencies:
    ```sh
    go mod tidy
    ```

3. Build the application:
    ```sh
    go build
    ```

## Usage

1. Run the application (without Docker):
    ```sh
    ./weather-app-jack-henry
    ```

2. The server will start on port 8080 by default. You can set a different port by setting the `PORT` environment variable:
    ```sh
    export PORT=8080
    ./weather-app-jack-henry
    ```

3. You can also run this service in Docker using:
    ```sh
    docker-compose up
    ```

## Endpoints

### Get Weather Data

- **URL:** `http://localhost:8080/weather?lat=<lat>&lon=<lon>`
- **Method:** `GET`
- **Query Parameters:**
    - `lat` (float): Latitude of the location.
    - `lon` (float): Longitude of the location.
- **Response:**
  ```json
  {
    "shortForecast": "Partly Cloudy",
    "temperature": 68,
    "temperatureUnit": "F",
    "weatherDescription": "moderate"
  }

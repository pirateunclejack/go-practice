package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type apiConfigData struct {
	OpenWeatherMapApiKey string `json:"OpenWeatherMapApiKey"`
}

type weatherData struct {
	Name string `json:"name"`
	Main struct {
		Kelvin float64 `json:"temp"`
	} `json:"main"`
}

func loadApiConfig(filename string) (apiConfigData, error){
	bytes, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("error reading api config file: %v", err)
		return apiConfigData{}, err
	}
	var c apiConfigData

	err = json.Unmarshal(bytes, &c)
	if err != nil {
		log.Printf("error unmarshal api config data: %v", err)
		return apiConfigData{}, err
	}

	return c, nil
}

func hello(w http.ResponseWriter, r *http.Request){
	log.Printf("Running hello handler...")
	w.Write([]byte("Hello from go\n"))
}

func query(city string)(weatherData, error) {
	apiConfig, err := loadApiConfig("./.apiConfig")
	if err != nil {
		log.Printf("error loading api config: %v", err)
		return weatherData{}, err
	}

	resp, err := http.Get("https://api.openweathermap.org/data/2.5/weather?APPID=" + apiConfig.OpenWeatherMapApiKey + "&q=" + city)
	if err != nil {
		log.Printf("error getting city temperature: %v", err)
		return weatherData{}, err
	}
	var d weatherData
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		log.Printf("error decoding weather data: %v", err)
		return weatherData{}, err
	}

	return d, nil
}

func getCityTemperature(w http.ResponseWriter, r *http.Request) {
	log.Printf("r.URL.Path: %v\n", r.URL.Path)
	city := strings.Split(r.URL.Path, "/")[2]
	log.Printf("city: %s\n", city)
	data, err := query(city)
	if err != nil {
		log.Printf("error querying city temperature: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}


func main() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/weather/", getCityTemperature)

	log.Printf("Starting weather tool...")
	http.ListenAndServe("localhost:8080", nil)
}

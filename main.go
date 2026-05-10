package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Dane autora i konfiguracja portu
const author = "Jakub Gwozdowski" 
const port = "8080"

type WeatherResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
	} `json:"current_weather"`
}

func main() {
	// HEALTHCHECK DLA WARSTWY SCRATCH 
	if len(os.Args) > 1 && os.Args[1] == "-health" {
		resp, err := http.Get("http://localhost:" + port + "/health")
		if err != nil || resp.StatusCode != 200 {
			os.Exit(1) 
		}
		os.Exit(0) 
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		html := `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Pogoda</title></head>
<body style="font-family: Arial, sans-serif; text-align: center; padding-top: 50px;">
	<h1>Sprawdź pogodę</h1>
	<form action="/weather" method="GET">
		<select name="coords" style="padding: 10px; font-size: 16px;">
			<option value="52.2297,21.0122">Warszawa, Polska</option>
			<option value="51.5085,-0.1257">Londyn, UK</option>
			<option value="40.7143,-74.006">Nowy Jork, USA</option>
			<option value="35.6895,139.6917">Tokio, Japonia</option>
		</select>
		<button type="submit" style="padding: 10px 20px; font-size: 16px;">Sprawdź</button>
	</form>
</body>
</html>`
		fmt.Fprint(w, html)
	})

	http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		coords := r.URL.Query().Get("coords")
		if coords == "" {
			http.Error(w, "Brak koordynatów", http.StatusBadRequest)
			return
		}

		parts := strings.Split(coords, ",")
		if len(parts) != 2 {
			http.Error(w, "Nieprawidłowe koordynaty", http.StatusBadRequest)
			return
		}

		// Zapytanie do API Open-Meteo (wymaga certyfikatów SSL w kontenerze)
		apiURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current_weather=true", parts[0], parts[1])
		resp, err := http.Get(apiURL)
		if err != nil {
			http.Error(w, "Błąd połączenia z API pogodowym", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var weather WeatherResponse
		if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
			http.Error(w, "Błąd parsowania danych API", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "<div style='font-family: Arial; text-align: center; padding-top: 50px;'>")
		fmt.Fprintf(w, "<h2>Aktualna temperatura: %.1f °C</h2>", weather.CurrentWeather.Temperature)
		fmt.Fprintf(w, "<br><a href='/' style='text-decoration: none; color: #0066cc;'>&larr; Wróć do wyboru</a></div>")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	
	dateStr := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] Aplikacja uruchomiona.\n", dateStr)
	fmt.Printf("Autor: %s\n", author)
	fmt.Printf("Serwer nasłuchuje na porcie TCP: %s\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
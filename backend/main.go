package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type ExchangeResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

type ExchangeRate struct {
	Base      string  `json:"base"`
	Target    string  `json:"target"`
	Rate      float64 `json:"rate"`
	Timestamp string  `json:"timestamp"`
}

var db *sql.DB

func getExchangeRate(base, target string) (float64, error) {
	url := fmt.Sprintf("https://api.exchangerate.host/latest?base=%s&symbols=%s", base, target)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data ExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	rate, exists := data.Rates[target]
	if !exists {
		return 0, fmt.Errorf("rate not found")
	}
	return rate, nil
}

func rateHandler(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	target := r.URL.Query().Get("target")
	if base == "" || target == "" {
		http.Error(w, "Missing base or target parameter", http.StatusBadRequest)
		return
	}

	rate, err := getExchangeRate(base, target)
	if err != nil {
		http.Error(w, "Failed to fetch rate", http.StatusInternalServerError)
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	_, err = db.Exec("INSERT INTO exchange_rates (base_currency, target_currency, rate, scraped_at) VALUES ($1, $2, $3, $4)", base, target, rate, timestamp)
	if err != nil {
		log.Println("Failed to insert into DB:", err)
	}

	json.NewEncoder(w).Encode(ExchangeRate{Base: base, Target: target, Rate: rate, Timestamp: timestamp})
}

func cronJob() {
	for {
		pairs := [][2]string{
			{"USD", "EUR"},
			{"USD", "JPY"},
			{"EUR", "GBP"},
		}
		for _, pair := range pairs {
			rate, err := getExchangeRate(pair[0], pair[1])
			if err != nil {
				log.Println("Cron fetch error:", err)
				continue
			}
			timestamp := time.Now().Format(time.RFC3339)
			_, err = db.Exec("INSERT INTO exchange_rates (base_currency, target_currency, rate, scraped_at) VALUES ($1, $2, $3, $4)", pair[0], pair[1], rate, timestamp)
			if err != nil {
				log.Println("Cron DB insert error:", err)
			}
		}
		time.Sleep(3600 * time.Second) // hourly
	}
}

func main() {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	go cronJob()

	http.HandleFunc("/api/latest", rateHandler)
	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
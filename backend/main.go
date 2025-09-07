package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Data structures
type Bar struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
}

type TechnicalIndicators struct {
	BollingerBands *BollingerBand `json:"bollinger_bands,omitempty"`
	MACD           *MACDData      `json:"macd,omitempty"`
	RSI            float64        `json:"rsi,omitempty"`
	SMA20          float64        `json:"sma_20,omitempty"`
	EMA20          float64        `json:"ema_20,omitempty"`
}

type BollingerBand struct {
	Upper  float64 `json:"upper"`
	Middle float64 `json:"middle"`
	Lower  float64 `json:"lower"`
}

type MACDData struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}

type CandlestickData struct {
	Bar        Bar                 `json:"bar"`
	Indicators TechnicalIndicators `json:"indicators"`
}

type ChartResponse struct {
	Symbol string            `json:"symbol"`
	Data   []CandlestickData `json:"data"`
}

// Alpaca API structures
type AlpacaBar struct {
	T string  `json:"t"` // timestamp
	O float64 `json:"o"` // open
	H float64 `json:"h"` // high
	L float64 `json:"l"` // low
	C float64 `json:"c"` // close
	V int64   `json:"v"` // volume
}

type AlpacaBarsResponse struct {
	Bars []AlpacaBar `json:"bars"`
}

// Configuration
var (
	alpacaAPIKey    = os.Getenv("APCA_API_KEY_ID")
	alpacaAPISecret = os.Getenv("APCA_API_SECRET_KEY")
	alpacaBaseURL   = "https://paper-api.alpaca.markets" // Paper trading endpoint
	alpacaDataURL   = "https://data.alpaca.markets/v2"   // ✅ v2 data API
)

// Technical indicator calculations
func calculateSMA(closes []float64, period int) float64 {
	if len(closes) < period {
		return 0
	}
	sum := 0.0
	for i := len(closes) - period; i < len(closes); i++ {
		sum += closes[i]
	}
	return sum / float64(period)
}

func calculateEMA(closes []float64, period int) float64 {
	if len(closes) < period {
		return 0
	}
	multiplier := 2.0 / float64(period+1)
	ema := closes[0]
	for i := 1; i < len(closes); i++ {
		ema = (closes[i] * multiplier) + (ema * (1 - multiplier))
	}
	return ema
}

func calculateBollingerBands(closes []float64, period int, stdDev float64) *BollingerBand {
	if len(closes) < period {
		return nil
	}
	sma := calculateSMA(closes, period)
	variance := 0.0
	for i := len(closes) - period; i < len(closes); i++ {
		variance += math.Pow(closes[i]-sma, 2)
	}
	stdDeviation := math.Sqrt(variance / float64(period))
	return &BollingerBand{
		Upper:  sma + (stdDev * stdDeviation),
		Middle: sma,
		Lower:  sma - (stdDev * stdDeviation),
	}
}

func calculateMACD(closes []float64) *MACDData {
	if len(closes) < 26 {
		return nil
	}
	ema12 := calculateEMA(closes, 12)
	ema26 := calculateEMA(closes, 26)
	macd := ema12 - ema26
	signal := calculateEMA([]float64{macd}, 9)
	histogram := macd - signal
	return &MACDData{
		MACD:      macd,
		Signal:    signal,
		Histogram: histogram,
	}
}

func calculateRSI(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 0
	}
	gains := make([]float64, 0)
	losses := make([]float64, 0)
	for i := 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, -change)
		}
	}
	if len(gains) < period {
		return 0
	}
	avgGain := calculateSMA(gains, period)
	avgLoss := calculateSMA(losses, period)
	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	return rsi
}

// API handlers
func getMarketData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	if symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}
	bars, err := fetchAlpacaBars(symbol, 100) // Get last 100 days
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %v", err), http.StatusInternalServerError)
		return
	}
	candlestickData := make([]CandlestickData, 0, len(bars))
	closes := make([]float64, 0, len(bars))
	sort.Slice(bars, func(i, j int) bool {
		return bars[i].Timestamp.Before(bars[j].Timestamp)
	})
	for _, bar := range bars {
		closes = append(closes, bar.Close)
	}
	for i, bar := range bars {
		indicators := TechnicalIndicators{}
		currentCloses := closes[:i+1]
		if len(currentCloses) >= 20 {
			indicators.SMA20 = calculateSMA(currentCloses, 20)
			indicators.EMA20 = calculateEMA(currentCloses, 20)
			indicators.BollingerBands = calculateBollingerBands(currentCloses, 20, 2)
		}
		if len(currentCloses) >= 26 {
			indicators.MACD = calculateMACD(currentCloses)
		}
		if len(currentCloses) >= 14 {
			indicators.RSI = calculateRSI(currentCloses, 14)
		}
		candlestickData = append(candlestickData, CandlestickData{
			Bar:        bar,
			Indicators: indicators,
		})
	}
	response := ChartResponse{
		Symbol: symbol,
		Data:   candlestickData,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func fetchAlpacaBars(symbol string, limit int) ([]Bar, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -limit).Format("2006-01-02")

	// ✅ Force IEX feed for free accounts
	url := fmt.Sprintf("%s/stocks/%s/bars?start=%s&end=%s&timeframe=1Day&limit=%d&feed=iex",
		alpacaDataURL, symbol, startDate, endDate, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("APCA-API-KEY-ID", alpacaAPIKey)
	req.Header.Set("APCA-API-SECRET-KEY", alpacaAPISecret)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var alpacaResponse AlpacaBarsResponse
	if err := json.NewDecoder(resp.Body).Decode(&alpacaResponse); err != nil {
		return nil, err
	}

	var bars []Bar
	for _, alpacaBar := range alpacaResponse.Bars {
		timestamp, err := time.Parse(time.RFC3339, alpacaBar.T)
		if err != nil {
			log.Printf("Error parsing timestamp %s: %v", alpacaBar.T, err)
			continue
		}
		bars = append(bars, Bar{
			Timestamp: timestamp,
			Open:      alpacaBar.O,
			High:      alpacaBar.H,
			Low:       alpacaBar.L,
			Close:     alpacaBar.C,
			Volume:    alpacaBar.V,
		})
	}
	return bars, nil
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func main() {
	if alpacaAPIKey == "" || alpacaAPISecret == "" {
		log.Fatal("APCA_API_KEY_ID and APCA_API_SECRET_KEY environment variables are required")
	}
	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", healthCheck).Methods("GET")
	api.HandleFunc("/chart/{symbol}", getMarketData).Methods("GET")

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000", "http://localhost:5173"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Using Alpaca paper trading environment (free IEX feed)")
	log.Fatal(http.ListenAndServe(":"+port, corsHandler))
}

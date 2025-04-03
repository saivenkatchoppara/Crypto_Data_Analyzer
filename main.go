package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Crypto struct to hold data
type Crypto struct {
	Name   string  `json:"name"`
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Change string  `json:"change"`
}

var cryptos []Crypto

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// Serve static CSS
	r.Static("/static", "./static")

	// Routes
	r.GET("/", showCryptos)
	r.GET("/api/cryptos", getCryptoData)
	r.GET("/api/cryptos/search", searchCrypto)
	r.GET("/api/cryptos/sort", sortCryptos)
	r.GET("/api/cryptos/top-gainer-loser", getTopGainerLoser)
	r.GET("/api/cryptos/download", downloadCSV)
	r.GET("/api/cryptos/fetch", fetchCryptoData)

	// Run server
	r.Run(":8080")
}

// Function to fetch crypto data from CoinGecko API
func fetchCryptoData(c *gin.Context) {
	// List of crypto symbols
	cryptoSymbols := []string{"bitcoin", "ethereum", "ripple", "solana", "dogecoin", "cardano", "shiba-inu"}

	url := "https://api.coingecko.com/api/v3/simple/price?ids=" + strings.Join(cryptoSymbols, ",") + "&vs_currencies=usd&include_24hr_change=true"

	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch crypto data"})
		return
	}
	defer resp.Body.Close()

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error decoding response"})
		return
	}

	// Clear existing cryptos
	cryptos = []Crypto{}

	// Parse and store data
	for symbol, data := range result {
		cryptos = append(cryptos, Crypto{
			Name:   strings.Title(symbol),
			Symbol: strings.ToUpper(symbol),
			Price:  data["usd"],
			Change: fmt.Sprintf("%.2f%%", data["usd_24h_change"]),
		})
	}

	// Save data to CSV
	saveToCSV()

	c.JSON(http.StatusOK, gin.H{"message": "Crypto data fetched successfully!"})
}

// Function to save crypto data to CSV
func saveToCSV() {
	file, err := os.Create("crypto_data.csv")
	if err != nil {
		fmt.Println("Error creating CSV file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Name", "Symbol", "Price (USD)", "24h Change"}
	writer.Write(headers)

	for _, crypto := range cryptos {
		writer.Write([]string{crypto.Name, crypto.Symbol, fmt.Sprintf("%.2f", crypto.Price), crypto.Change})
	}

	fmt.Println("Crypto data successfully saved to CSV!")
}

// API to return all crypto data
func getCryptoData(c *gin.Context) {
	c.JSON(http.StatusOK, cryptos)
}

// API to search for a specific crypto by symbol
func searchCrypto(c *gin.Context) {
	query := strings.ToUpper(c.Query("symbol"))

	for _, crypto := range cryptos {
		if crypto.Symbol == query {
			c.JSON(http.StatusOK, crypto)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Crypto not found"})
}

// API to sort cryptos by price (ascending or descending)
func sortCryptos(c *gin.Context) {
	sortOrder := c.DefaultQuery("order", "asc")

	if sortOrder == "asc" {
		sort.Slice(cryptos, func(i, j int) bool {
			return cryptos[i].Price < cryptos[j].Price
		})
	} else {
		sort.Slice(cryptos, func(i, j int) bool {
			return cryptos[i].Price > cryptos[j].Price
		})
	}

	c.JSON(http.StatusOK, cryptos)
}

// API to get the top gainer and loser
func getTopGainerLoser(c *gin.Context) {
	if len(cryptos) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No crypto data available"})
		return
	}

	var topGainer, topLoser Crypto
	topGainerChange := -100.0
	topLoserChange := 100.0

	for _, crypto := range cryptos {
		changeStr := strings.Trim(crypto.Change, " %")
		changePercentage, err := strconv.ParseFloat(changeStr, 64)
		if err != nil {
			continue
		}

		if changePercentage > topGainerChange {
			topGainerChange = changePercentage
			topGainer = crypto
		}

		if changePercentage < topLoserChange {
			topLoserChange = changePercentage
			topLoser = crypto
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"top_gainer": topGainer,
		"top_loser":  topLoser,
	})
}

// API to download crypto data as CSV
func downloadCSV(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=crypto_data.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	writer.Write([]string{"Name", "Symbol", "Price (USD)", "24h Change"})

	for _, crypto := range cryptos {
		writer.Write([]string{crypto.Name, crypto.Symbol, fmt.Sprintf("%.2f", crypto.Price), crypto.Change})
	}
}

// HTML page to show cryptos
func showCryptos(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"cryptos": cryptos,
	})
}

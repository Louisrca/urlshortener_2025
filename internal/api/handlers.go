package api

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/axellelanca/urlshortener/internal/config"
	"github.com/axellelanca/urlshortener/internal/models"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Déclaré ici, initialisé dans run-server
var ClickEventsChannel chan models.ClickEvent

// ----------------------------
// ROUTES
// ----------------------------
func SetupRoutes(router *gin.Engine, linkService *services.LinkService) {

	// Health check
	router.GET("/health", HealthCheckHandler)

	// Routes API versionnées
	api := router.Group("/api/v1")
	{
		api.POST("/links", CreateShortLinkHandler(linkService))
		api.GET("/links/:shortCode/stats", GetLinkStatsHandler(linkService))
	}

	// Redirection short URL
	router.GET("/:shortCode", RedirectHandler(linkService))
}

// Healthcheck simple
func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DTO
type CreateLinkRequest struct {
	LongURL string `json:"long_url" binding:"required,url"`
}

// Handler création d'un lien court
func CreateShortLinkHandler(linkService *services.LinkService) gin.HandlerFunc {

	cfg, _ := config.LoadConfig()

	return func(c *gin.Context) {
		var req CreateLinkRequest

		// Validation JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Appel du service
		link, err := linkService.CreateLink(req.LongURL)
		if err != nil {
			log.Printf("Error creating short link for %s: %v", req.LongURL, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"short_code":     link.ShortCode,
			"long_url":       link.LongURL,
			"full_short_url": cfg.Server.BaseURL + "/" + link.ShortCode,
		})
	}
}

// Handler redirection
func RedirectHandler(linkService *services.LinkService) gin.HandlerFunc {
	return func(c *gin.Context) {

		shortCode := c.Param("shortCode")

		link, err := linkService.GetLinkByShortCode(shortCode)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
				return
			}

			log.Printf("Error retrieving link for %s: %v", shortCode, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Construire l’événement de clic
		clickEvent := models.ClickEvent{
			LinkID:    link.ID,
			Timestamp: time.Now(),
			UserAgent: c.GetHeader("User-Agent"),
			IPAddress: c.ClientIP(),
		}

		// Envoi non bloquant dans le channel
		select {
		case ClickEventsChannel <- clickEvent:
		default:
			log.Printf("Warning: ClickEventsChannel is full, dropping event for %s", shortCode)
		}

		// Redirection 302 vers l'URL longue
		c.Redirect(http.StatusFound, link.LongURL)
	}
}

// Handler stats
func GetLinkStatsHandler(linkService *services.LinkService) gin.HandlerFunc {
	return func(c *gin.Context) {

		shortCode := c.Param("shortCode")

		link, totalClicks, err := linkService.GetLinkStats(shortCode)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
				return
			}

			log.Printf("Error retrieving stats for %s: %v", shortCode, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"short_code":   link.ShortCode,
			"long_url":     link.LongURL,
			"total_clicks": totalClicks,
		})
	}
}

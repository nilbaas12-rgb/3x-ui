package controller

import (
	"net/http"
	"strconv"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service"
	"github.com/gin-gonic/gin"
)

// AWG2Controller handles AWG2-specific API endpoints
type AWG2Controller struct {
	awg2Service service.AWG2Service
}

// NewAWG2Controller creates a new AWG2Controller
func NewAWG2Controller(g *gin.RouterGroup, awg2Service *service.AWG2Service) *AWG2Controller {
	a := &AWG2Controller{
		awg2Service: *awg2Service,
	}
	a.initRouter(g)
	return a
}

// initRouter sets up the AWG2 routes
func (a *AWG2Controller) initRouter(g *gin.RouterGroup) {
	g = g.Group("/awg2")

	// Inbound operations
	g.GET("/inbounds/:id", a.getInbound)
	g.POST("/inbounds/:id/randomize", a.randomizeParams)

	// Client operations
	g.POST("/inbounds/:id/client/add", a.addClient)
	g.GET("/inbounds/:id/clients", a.listClients)
	g.POST("/client/:id/remove", a.removeClient)
	g.GET("/client/:id/config", a.getClientConfig)
	g.GET("/client/:id/qr", a.getClientQR)
	g.GET("/client/:id/traffic", a.getClientTraffic)

	// Stats
	g.GET("/inbounds/:id/stats", a.getStats)
}

// getInbound retrieves an AWG2 inbound
func (a *AWG2Controller) getInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inbound ID"})
		return
	}

	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"status": "ok", "inboundId": id})
}

// randomizeParams randomizes obfuscation parameters
func (a *AWG2Controller) randomizeParams(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inbound ID"})
		return
	}

	inbound, err := a.awg2Service.RandomizeParams(id)
	if err != nil {
		logger.Error("[awg2] Randomize failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inbound)
}

// addClient adds a new client
func (a *AWG2Controller) addClient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inbound ID"})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required"`
		Name  string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := a.awg2Service.AddClient(id, req.Email, req.Name)
	if err != nil {
		logger.Error("[awg2] Add client failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

// listClients retrieves all clients for an inbound
func (a *AWG2Controller) listClients(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inbound ID"})
		return
	}

	clients, err := a.awg2Service.ListClients(id)
	if err != nil {
		logger.Error("[awg2] List clients failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, clients)
}

// removeClient removes a client
func (a *AWG2Controller) removeClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client ID"})
		return
	}

	err = a.awg2Service.RemoveClient(uint(id))
	if err != nil {
		logger.Error("[awg2] Remove client failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// getClientConfig returns the client configuration
func (a *AWG2Controller) getClientConfig(c *gin.Context) {
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client ID"})
		return
	}

	inboundID := c.Query("inbound")
	endpoint := c.Query("endpoint")

	if inboundID == "" || endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing inbound or endpoint"})
		return
	}

	id, _ := strconv.Atoi(inboundID)

	config, err := a.awg2Service.GetClientConfig(id, uint(clientID), endpoint)
	if err != nil {
		logger.Error("[awg2] Get config failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", "attachment; filename=client.conf")
	c.String(http.StatusOK, config)
}

// getClientQR returns the QR code for a client
func (a *AWG2Controller) getClientQR(c *gin.Context) {
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client ID"})
		return
	}

	inboundID := c.Query("inbound")
	endpoint := c.Query("endpoint")

	if inboundID == "" || endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing inbound or endpoint"})
		return
	}

	id, _ := strconv.Atoi(inboundID)

	qrData, err := a.awg2Service.GetClientQR(id, uint(clientID), endpoint)
	if err != nil {
		logger.Error("[awg2] Get QR failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", "inline; filename=client.png")
	c.Data(http.StatusOK, "image/png", qrData)
}

// getClientTraffic returns traffic statistics for a client
func (a *AWG2Controller) getClientTraffic(c *gin.Context) {
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client ID"})
		return
	}

	// TODO: Implement traffic retrieval
	c.JSON(http.StatusOK, gin.H{"id": clientID, "download": 0, "upload": 0})
}

// getStats returns statistics for an inbound
func (a *AWG2Controller) getStats(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inbound ID"})
		return
	}

	rx, tx, err := a.awg2Service.GetStats(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inboundId": id,
		"download": rx,
		"upload":   tx,
	})
}

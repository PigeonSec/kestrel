package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pigeonsec/kestrel/internal/misp"
)

// MISPHandler handles MISP feed endpoints
type MISPHandler struct {
	misp *misp.Handler
}

// NewMISPHandler creates a new MISP handler
func NewMISPHandler(mispHandler *misp.Handler) *MISPHandler {
	return &MISPHandler{
		misp: mispHandler,
	}
}

// GetManifest handles GET /misp/manifest.json
func (h *MISPHandler) GetManifest(c *gin.Context) {
	data := h.misp.GetManifest()
	c.Data(http.StatusOK, "application/json", data)
}

// GetEvent handles GET /misp/events/:id.json
func (h *MISPHandler) GetEvent(c *gin.Context) {
	eventID := c.Param("id")

	data, err := h.misp.GetEvent(c.Request.Context(), eventID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}

// GetAllEvents handles GET /misp/events
func (h *MISPHandler) GetAllEvents(c *gin.Context) {
	data := h.misp.GetAllEvents(c.Request.Context())
	c.Data(http.StatusOK, "application/json", data)
}

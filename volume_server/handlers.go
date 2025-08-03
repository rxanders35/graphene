package volume_server

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VolumeHandler struct {
	volume *Volume
}

func (v *VolumeHandler) Write(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid req body"})
		return
	}
	needleId, err := v.volume.Write(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": needleId.String()})
}

func (v *VolumeHandler) Read(c *gin.Context) {
	uuidStr := c.Param("uuid")
	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid req body"})
		return
	}

	data, err := v.volume.Read(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read"})
		return
	}
	c.Data(http.StatusOK, "application/octect-stream", data)
}

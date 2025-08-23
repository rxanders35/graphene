package volume_server

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VolumeHandler struct {
	storage StorageEngine
}

func NewVolumeHandler(s StorageEngine) *VolumeHandler {
	return &VolumeHandler{
		storage: s,
	}
}

func (v *VolumeHandler) Write(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid req body"})
		return
	}
	needleId := uuid.New()
	err = v.storage.Write(needleId, data)
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

	data, err := v.storage.Read(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read"})
		return
	}
	c.Data(http.StatusOK, "application/octet-stream", data)
}

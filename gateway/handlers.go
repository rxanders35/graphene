package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pb "github.com/rxanders35/graphene/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GatewayHandler struct {
	masterClient *MasterClient
	httpClient   *http.Client
}

func NewGatewayHandler(m *MasterClient) (*GatewayHandler, error) {
	g := &GatewayHandler{
		masterClient: m,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	return g, nil
}

func (g *GatewayHandler) Write(c *gin.Context) {
	masterReq := &pb.AssignVolumeRequest{}
	masterResp, err := g.masterClient.client.AssignVolume(c, masterReq)
	if err != nil {
		log.Printf("Failed to get a volume from the master: %v", err)
		st, _ := status.FromError(err)
		if st.Code() == codes.Unavailable {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no storage volumes available"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "master server internal error"})
		return
	}

	volumeId, err := uuid.FromBytes(masterResp.GetVolumeId())
	if err != nil {
		log.Printf("Master returned invalid volume UUID bytes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "master server returned invalid data"})
		return
	}

	volumeAddr := fmt.Sprintf("http://%s/v1/volume/write", masterResp.HttpAddress)
	volumeReq, err := http.NewRequest("POST", volumeAddr, c.Request.Body)
	if err != nil {
		log.Printf("Failed to build post req for volume server: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	volumeReq.Header.Set("Content-Type", c.GetHeader("Content-Type"))

	volumeResp, err := g.httpClient.Do(volumeReq)
	if err != nil {
		log.Printf("Failed to send data to volume %s: %v", volumeId, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not write to volume server"})
		return
	}
	defer volumeResp.Body.Close()

	if volumeResp.StatusCode != http.StatusCreated {
		log.Printf("Volume server returned non-201 status: %d", volumeResp.StatusCode)
		c.JSON(http.StatusBadGateway, gin.H{"error": "volume server failed to store data"})
		return
	}

	body, err := io.ReadAll(volumeResp.Body)
	if err != nil {
		log.Printf("Failed to read volume server resp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var respData struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &respData); err != nil {
		log.Printf("Failed to unmarshal JSON from volume server: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response from volume server"})
		return
	}

	fatID := fmt.Sprintf("%s:%s", volumeId.String(), respData.ID)
	c.JSON(http.StatusCreated, gin.H{"id": fatID})
}

func (g *GatewayHandler) Read(c *gin.Context) {
	fatID := c.Param("fat_id")

	parts := strings.Split(fatID, ":")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid object id format"})
		return
	}

	volumeIdStr, needleIdStr := parts[0], parts[1]
	volumeId, err := uuid.Parse(volumeIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid volume id format"})
		return
	}
	if _, err := uuid.Parse(needleIdStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid needle id format"})
		return
	}

	masterReq := &pb.GetVolumeLocationRequest{VolumeId: volumeId[:]}
	masterResp, err := g.masterClient.client.GetVolumeLocation(c, masterReq)
	if err != nil {
		log.Printf("Failed to get volume location from master: %v", err)
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "master server is unavailable"})
		return
	}

	volumeAddr := fmt.Sprintf("http://%s/v1/volume/read/%s", masterResp.HttpAddress, needleIdStr)
	volumeResp, err := g.httpClient.Get(volumeAddr)
	if err != nil {
		log.Printf("Failed to get data from volume %s: %v", volumeId, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not read from volume server"})
		return
	}
	defer volumeResp.Body.Close()

	c.DataFromReader(volumeResp.StatusCode, volumeResp.ContentLength, volumeResp.Header.Get("Content-Type"), volumeResp.Body, nil)
}

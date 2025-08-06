package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pb "github.com/rxanders35/sss/proto"
)

type GatewayHandler struct {
	masterClient *MasterClient
}

func NewGatewayHandler(m *MasterClient) (*GatewayHandler, error) {
	g := &GatewayHandler{
		masterClient: m,
	}
	// figure it out
	return g, nil
}

func (g *GatewayHandler) Write(c *gin.Context) {
	masterReq := &pb.AssignVolumeRequest{}
	masterResp, err := g.masterClient.client.AssignVolume(c, masterReq)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "failed to get a volume from the master"})
		return
	}
	client := &http.Client{} //inject into handler later
	volumeAddr := fmt.Sprintf("http://%s/v1/volume/write", masterResp.HttpAddress)
	volumeReq, err := http.NewRequest("POST", volumeAddr, c.Request.Body)
	if err != nil {
		log.Printf("Failed to build post req for sending data to volume. Why: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build post req for sending data to volume"})
		return
	}
	volumeResp, err := client.Do(volumeReq)
	if err != nil {
		log.Printf("Failed to send data to volume. Why: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write to volume"})
		return
	}
	defer volumeResp.Body.Close()

	body, err := io.ReadAll(volumeResp.Body)
	if err != nil {
		log.Printf("Failed to read volume server resp. Why: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to volume server resp"})
		return
	}

	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Printf("Failed to extract response from volume server. Why: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to extract response from volume server"})
		return
	}

	id, err := uuid.Parse(resp.ID)
	if err != nil {
		log.Printf("Failed to extract id from volume server resp. Why: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to extract id from volume server resp"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": id})
}

func (g *GatewayHandler) Read(c *gin.Context) {
	req := &pb.AssignVolumeRequest{}
	resp, err := g.masterClient.client.AssignVolume(c, req)
	if err != nil {
		log.Printf("Failed to get a volume %s with addr %s ", resp.HttpAddress, resp.VolumeId)
	}

	client := &http.Client{}

	//todo: impl read path
}

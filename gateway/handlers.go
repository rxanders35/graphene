package gateway

import "github.com/gin-gonic/gin"

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

}
func (g *GatewayHandler) Read(c *gin.Context) {

}

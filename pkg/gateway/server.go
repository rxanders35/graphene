package gateway

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GatewayServer struct {
	addr           string
	engine         *gin.Engine
	srv            *http.Server
	gatewayHandler *GatewayHandler
	masterClient   *MasterClient
}

func NewGatewayServer(gatewayAddr string, h *GatewayHandler) (*GatewayServer, error) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	g := &GatewayServer{
		addr:           gatewayAddr,
		engine:         engine,
		gatewayHandler: h,
	}
	g.registerRoutes()

	return g, nil
}

func (g *GatewayServer) registerRoutes() {
	v1 := g.engine.Group("/v1")

	gateway := v1.Group("/gateway")

	gateway.POST("/write", g.gatewayHandler.Write)
	// Encapsulates the entire write flow (req Master for volume addr -> forward to volume server)
	gateway.GET("/read/:fat_id", g.gatewayHandler.Read)
	// Encapsulates the entire read flow (parse fat_id -> req Master for volume addr -> forward to volume server)
}

func (g *GatewayServer) Run() error {
	g.srv = &http.Server{
		Addr:    g.addr,
		Handler: g.engine,
	}
	return g.srv.ListenAndServe()
}

func (g *GatewayServer) Shutdown(ctx context.Context) error {
	if g.srv == nil {
		return nil
	}
	return g.srv.Shutdown(ctx)
}

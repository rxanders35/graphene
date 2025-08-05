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
	gatewayHandler GatewayHandler
	masterClient   *MasterClient
}

func NewGatewayServer(gatewayAddr, masterGRPCaddr string) (*GatewayServer, error) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	handler := GatewayHandler{} //todo

	/*
		c, err := NewGRPCclient(masterGRPCaddr)
		if err != nil {
			log.Printf("Failed to start gRPC client. Why: %v", err)
		}
	*/

	g := &GatewayServer{
		addr:           gatewayAddr,
		engine:         engine,
		gatewayHandler: handler,
		// grpcClient: c,
	}
	g.registerRoutes()

	// figure out RPC for registering gateway
	// c.RegisterGateway()

	return g, nil
}

func (g *GatewayServer) registerRoutes() {
	v1 := g.engine.Group("/v1")

	volume := v1.Group("/gateway")

	volume.POST("/write", g.gatewayHandler.Write)
	//Encapsulates the entire write flow (req Master for volume addr -> use the volume addr to call the volume server's /write -> Write())
	volume.GET("/read/:object_name", g.gatewayHandler.Read) //need to abrtract uuid from user (object_name -> object's uuid)
	//Encapsulates the entire read flow (req Master for volume addr -> use the volume addr to call the volume server's /read/:uuid-> Read())
}

func (g *GatewayServer) Run() error {
	g.srv = &http.Server{
		Addr:    g.addr,
		Handler: g.engine,
	}
	return g.srv.ListenAndServe()
}

func (h *GatewayServer) Shutdown(ctx context.Context) error {
	if h.srv == nil {
		return nil
	}
	return h.srv.Shutdown(ctx)
}

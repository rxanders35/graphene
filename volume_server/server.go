package volume_server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Server struct {
	HTTPServer HTTPServer
}

type HTTPServer struct {
	addr          string
	volume        *Volume
	engine        *gin.Engine
	volumeHandler VolumeHandler
	srv           *http.Server
	grpcClient    *MasterClient
}

func NewHTTPServer(volumeHTTPaddr, masterGRPCaddr string, v *Volume, volServerID uuid.UUID, m *MasterClient) (*HTTPServer, error) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	handler := VolumeHandler{volume: v}

	h := &HTTPServer{
		addr:          volumeHTTPaddr,
		volume:        v,
		engine:        engine,
		volumeHandler: handler,
		grpcClient:    m,
	}
	h.registerRoutes()

	return h, nil
}

func (h *HTTPServer) registerRoutes() {
	v1 := h.engine.Group("/v1")

	volume := v1.Group("/volume")

	volume.POST("/write", h.volumeHandler.Write)
	volume.GET("/read/:uuid", h.volumeHandler.Read)
}

func (h *HTTPServer) Run() error {
	h.srv = &http.Server{
		Addr:    h.addr,
		Handler: h.engine,
	}
	return h.srv.ListenAndServe()
}

func (h *HTTPServer) Shutdown(ctx context.Context) error {
	if h.srv == nil {
		return nil
	}
	return h.srv.Shutdown(ctx)
}

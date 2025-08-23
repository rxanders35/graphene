package volume_server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pb "github.com/rxanders35/graphene/proto"
	"google.golang.org/grpc"
)

type Server struct {
	HTTPServer HTTPServer
}

type HTTPServer struct {
	volumeHTTPaddr string
	engine         *gin.Engine
	handler        *VolumeHandler
	srv            *http.Server
	grpcClient     *MasterClient
}

func NewHTTPServer(v string, s StorageEngine, m *MasterClient, volSrvID uuid.UUID) (*HTTPServer, error) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	handler := NewVolumeHandler(s)

	h := &HTTPServer{
		volumeHTTPaddr: v,
		engine:         engine,
		handler:        handler,
		grpcClient:     m,
	}
	h.registerRoutes()

	req := &pb.RegisterVolumeRequest{
		HttpAddress: v,
		VolumeId:    volSrvID[:],
	}

	m.Client.RegisterVolume(context.Background(), req, grpc.WaitForReady(true))

	return h, nil
}

func (h *HTTPServer) registerRoutes() {
	v1 := h.engine.Group("/v1")

	volume := v1.Group("/volume")

	volume.POST("/write", h.handler.Write)
	volume.GET("/read/:uuid", h.handler.Read)
}

func (h *HTTPServer) Run() error {
	h.srv = &http.Server{
		Addr:    h.volumeHTTPaddr,
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

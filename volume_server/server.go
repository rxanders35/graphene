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
	addr          string
	volume        *NeedleVolume
	engine        *gin.Engine
	volumeHandler VolumeHandler
	srv           *http.Server
	grpcClient    *MasterClient
}

func NewHTTPServer(volumeHTTPaddr string, v *NeedleVolume, m *MasterClient, volServerID uuid.UUID) (*HTTPServer, error) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	handler := VolumeHandler{needleVolume: v}

	h := &HTTPServer{
		addr:          volumeHTTPaddr,
		volume:        v,
		engine:        engine,
		volumeHandler: handler,
		grpcClient:    m,
	}
	h.registerRoutes()

	req := &pb.RegisterVolumeRequest{
		HttpAddress: volumeHTTPaddr,
		VolumeId:    volServerID[:],
	}

	m.Client.RegisterVolume(context.Background(), req, grpc.WaitForReady(true))

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

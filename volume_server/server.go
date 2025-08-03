package volume_server

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pb "github.com/rxanders35/sss/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	grpcClient    *GRPCclient
}

func NewHTTPServer(volumeHTTPaddr, masterGRPCaddr string, v *Volume, volServerID uuid.UUID) (*HTTPServer, error) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	handler := VolumeHandler{volume: v}
	h := &HTTPServer{
		addr:          volumeHTTPaddr,
		volume:        v,
		engine:        engine,
		volumeHandler: handler,
	}
	h.registerRoutes()

	g, err := NewGRPCclient(masterGRPCaddr)
	if err != nil {
		log.Printf("Failed to start gRPC client. Why: %v", err)
	}

	req := &pb.RegisterVolumeRequest{
		HttpAddress: volumeHTTPaddr,
		VolumeId:    volServerID[:],
	}

	g.client.RegisterVolume(context.Background(), req)

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

type GRPCclient struct {
	masterAddr string
	conn       *grpc.ClientConn
	client     pb.MasterServiceClient
}

func NewGRPCclient(m string) (*GRPCclient, error) {
	conn, err := grpc.NewClient(m, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed dialing master. Why: %v", err)
	}

	client := pb.NewMasterServiceClient(conn)
	return &GRPCclient{
		masterAddr: m,
		conn:       conn,
		client:     client,
	}, nil
}

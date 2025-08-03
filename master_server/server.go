package master_server

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/google/uuid"
	pb "github.com/rxanders35/sss/proto"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	port          string
	volumeServers map[uuid.UUID]string //volume server id -> addr
	srv           *grpc.Server
	mu            sync.Mutex
	pb.UnimplementedMasterServiceServer
}

func NewGRPCServer(port string) *GRPCServer {
	volumeServers := make(map[uuid.UUID]string)

	s := grpc.NewServer()
	g := &GRPCServer{
		port:          port,
		volumeServers: volumeServers,
		srv:           s,
	}

	pb.RegisterMasterServiceServer(s, g)

	return g
}

func (g *GRPCServer) Run() {
	listener, err := net.Listen("tcp", g.port)
	if err != nil {
		log.Printf("Failed to init tcp listener on port: %s. Why: %v", g.port, err)
	}

	if err := g.srv.Serve(listener); err != nil {
		log.Printf("Failed to init gRPC server on top of tcp listener. Why %v", err)
	}
}

func (g *GRPCServer) RegisterVolume(ctx context.Context, req *pb.RegisterVolumeRequest) (*pb.RegisterVolumeResponse, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	volumeId, volumeAddr := req.GetVolumeId(), req.GetHttpAddress()
	var uuid uuid.UUID
	copy(uuid[:], volumeId)

	g.volumeServers[uuid] = volumeAddr

	return &pb.RegisterVolumeResponse{}, nil
}

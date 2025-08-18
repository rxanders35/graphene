package master_server

import (
	"context"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/rxanders35/sss/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	addr          string
	volumeServers map[uuid.UUID]string // volume server id -> addr
	srv           *grpc.Server
	mu            sync.RWMutex
	rand          *rand.Rand
	pb.UnimplementedMasterServiceServer
}

func NewGRPCServer(addr string) *GRPCServer {
	volumeServers := make(map[uuid.UUID]string)

	s := grpc.NewServer()
	g := &GRPCServer{
		addr:          addr,
		volumeServers: volumeServers,
		srv:           s,
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	pb.RegisterMasterServiceServer(s, g)

	return g
}

func (g *GRPCServer) Run() {
	listener, err := net.Listen("tcp", g.addr)
	if err != nil {
		log.Fatalf("Failed to init tcp listener on addr: %s. Why: %v", g.addr, err)
	}

	log.Printf("Master server listening on %s", g.addr)
	if err := g.srv.Serve(listener); err != nil {
		log.Fatalf("Failed to init gRPC server on top of tcp listener. Why %v", err)
	}
}

func (g *GRPCServer) RegisterVolume(ctx context.Context, req *pb.RegisterVolumeRequest) (*pb.RegisterVolumeResponse, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	volumeIdBytes, volumeAddr := req.GetVolumeId(), req.GetHttpAddress()
	volumeId, err := uuid.FromBytes(volumeIdBytes)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid volume id format")
	}

	g.volumeServers[volumeId] = volumeAddr
	log.Printf("Volume %s at addr %s successfully registered", volumeId, volumeAddr)

	return &pb.RegisterVolumeResponse{}, nil
}

func (g *GRPCServer) AssignVolume(ctx context.Context, req *pb.AssignVolumeRequest) (*pb.AssignVolumeResponse, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.volumeServers) == 0 {
		return nil, status.Errorf(codes.Unavailable, "no volume servers available")
	}

	keys := make([]uuid.UUID, 0, len(g.volumeServers))
	for k := range g.volumeServers {
		keys = append(keys, k)
	}

	randomKey := keys[g.rand.Intn(len(keys))]
	addr := g.volumeServers[randomKey]

	return &pb.AssignVolumeResponse{
		HttpAddress: addr,
		VolumeId:    randomKey[:],
	}, nil
}

func (g *GRPCServer) GetVolumeLocation(ctx context.Context, req *pb.GetVolumeLocationRequest) (*pb.GetVolumeLocationResponse, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	volumeIdBytes := req.GetVolumeId()
	volumeId, err := uuid.FromBytes(volumeIdBytes)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid volume id format")
	}

	addr, ok := g.volumeServers[volumeId]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "volume id not found: %s", volumeId)
	}

	return &pb.GetVolumeLocationResponse{
		HttpAddress: addr,
	}, nil
}

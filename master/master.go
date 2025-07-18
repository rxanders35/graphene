package master

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/rxanders35/sss/cluster"
	"google.golang.org/grpc"
)

type Master struct {
	cluster.UnimplementedClusterServiceServer
	liveServers     map[string]*ServerInfo
	volumeLocations map[uuid.UUID]string
	mu              sync.Mutex
}

type ServerInfo struct{}

func NewMaster() *Master {
	return &Master{
		liveServers:     make(map[string]*ServerInfo),
		volumeLocations: make(map[uuid.UUID]string),
	}
}

func (m *Master) Start(grpcAddr string) error {
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("Failed to start TCP listener on Master startup: %s", err)
	}
	grpcServer := grpc.NewServer()
	cluster.RegisterClusterServiceServer(grpcServer, m)

	err = grpcServer.Serve(listener)
	if err != nil {
		return fmt.Errorf("Failed to start GRPC server on Master startup: %s", err)
	}
	return nil
}

func (m *Master) Register(ctx context.Context, req *cluster.RegisterRequest) (*cluster.RegisterResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("Registered new volume server at: %s", req.HttpAddr)

	m.liveServers[req.HttpAddr] = &ServerInfo{}
	return &cluster.RegisterResponse{Success: true}, nil
}

func (m *Master) ReportStore(ctx context.Context, req *cluster.ReportStoreRequest) (*cluster.ReportStoreResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("")

	m.volumeLocations[uuid.UUID(req.ObjectId)] = req.VolumeServerAddr
	return &cluster.ReportStoreResponse{Success: true}, nil
}

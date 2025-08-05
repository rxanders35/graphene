package volume_server

import (
	"log"
	"sync"

	pb "github.com/rxanders35/sss/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MasterClient struct {
	masterAddr string
	conn       *grpc.ClientConn
	Client     pb.MasterServiceClient
}

// client singleton
func NewMasterClient(m string) *MasterClient {
	var (
		instance *MasterClient
		once     sync.Once
	)
	once.Do(func() {
		conn, err := grpc.NewClient(m, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed dialing master. Why: %v", err)
		}

		instance = &MasterClient{
			masterAddr: m,
			conn:       conn,
			Client:     pb.NewMasterServiceClient(conn),
		}
	})
	return instance
}

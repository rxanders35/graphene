package gateway

import (
	"log"
	"sync"

	pb "github.com/rxanders35/sss/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	instance *MasterClient
	once     sync.Once
)

type MasterClient struct {
	masterAddr string
	conn       *grpc.ClientConn
	client     pb.MasterServiceClient
}

func NewMasterclient(m string) *MasterClient {
	once.Do(func() {
		conn, err := grpc.NewClient(m, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed dialing master. Why: %v", err)
		}

		client := pb.NewMasterServiceClient(conn)
		instance = &MasterClient{
			masterAddr: m,
			conn:       conn,
			client:     client,
		}
	})
	return instance

}

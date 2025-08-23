package volume_server

import (
	"log"

	pb "github.com/rxanders35/graphene/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MasterClient struct {
	masterAddr string
	conn       *grpc.ClientConn
	Client     pb.MasterServiceClient
}

func NewMasterClient(m string) (*MasterClient, error) {
	conn, err := grpc.NewClient(m, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed dialing master. Why: %v", err)
	}

	c := &MasterClient{
		masterAddr: m,
		conn:       conn,
		Client:     pb.NewMasterServiceClient(conn),
	}

	return c, nil
}

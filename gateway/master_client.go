package gateway

import (
	"log"

	pb "github.com/rxanders35/sss/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MasterClient struct {
	masterAddr string
	conn       *grpc.ClientConn
	client     pb.MasterServiceClient
}

func NewMasterclient(masterAddr string) (*MasterClient, error) {
	conn, err := grpc.NewClient(masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed dialing master. Why: %v", err)
	}

	c := &MasterClient{
		masterAddr: masterAddr,
		conn:       conn,
		client:     pb.NewMasterServiceClient(conn),
	}
	return c, nil
}

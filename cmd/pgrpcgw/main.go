package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/the-maldridge/popcorn/internal/gateway"

	pb "github.com/the-maldridge/popcorn/internal/proto"
)

func main() {
	log.Println("Starting popcorn grpc gateway")

	l, err := net.Listen("tcp", os.Getenv("BIND"))
	if err != nil {
		log.Fatal(err)
	}

	var opts []grpc.ServerOption
	srvr := grpc.NewServer(opts...)
	pb.RegisterPopCornServer(srvr, gateway.New())
	srvr.Serve(l)
}

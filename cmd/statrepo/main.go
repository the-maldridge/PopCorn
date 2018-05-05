package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/the-maldridge/popcorn/internal/proto"
	"github.com/the-maldridge/popcorn/internal/repo"
)

var (
	addr = flag.String("addr", "", "Address to bind on")
	port = flag.Int("port", 8080, "Port to bind on")
)

func main() {
	flag.Parse()

	log.Println("Starting the stats repo")

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *addr, *port))
	if err != nil {
		log.Fatal(err)
	}

	var opts []grpc.ServerOption
	srvr := grpc.NewServer(opts...)
	pb.RegisterPopCornServer(srvr, &repo.StatsRepo{})
	srvr.Serve(l)
}

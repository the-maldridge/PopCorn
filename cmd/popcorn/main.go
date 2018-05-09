package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	pb "github.com/the-maldridge/popcorn/internal/proto"
)

var (
	uuidPath = flag.String("uuid_path", "/etc/popcorn/uuid", "Path to the uuid file")

	server = flag.String("server", "localhost", "Server to send stats to")
	port   = flag.Int("port", 8080, "Port to use on the server")

	statInterval = flag.Duration("interval", 24*time.Hour, "Interval to send stats on")

	machineID = ""
)

func getUUID() string {
	ID, err := ioutil.ReadFile(*uuidPath)
	if os.IsNotExist(err) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		ID = make([]byte, 96)
		r.Read(ID)
		if err := ioutil.WriteFile(*uuidPath, ID, 0644); err != nil {
			log.Fatal(err)
		}
	}
	h := sha256.Sum256(ID)
	return fmt.Sprintf("%x", h)
}

func getPkgs() []*pb.Package {
	_, err := exec.LookPath("xbps-query")
	if err != nil {
		log.Println("xbps-query isn't in $PATH.  Are you sure this is a Void system?")
		log.Fatal(err)
	}

	xbpsQueryCmd := exec.Command("xbps-query", "-m")
	var out bytes.Buffer
	xbpsQueryCmd.Stdout = &out
	if err := xbpsQueryCmd.Run(); err != nil {
		log.Fatal(err)
	}

	pkgs := []*pb.Package{}
	for _, p := range strings.Split(out.String(), "\n") {
		parts := strings.Split(p, "-")
		pkg := pb.Package{
			Name:    proto.String(strings.Join(parts[:len(parts)-1], "-")),
			Version: proto.String(parts[len(parts)-1]),
		}
		if pkg.GetName() == "" {
			continue
		}
		pkgs = append(pkgs, &pkg)
	}
	return pkgs
}

func getXUname() *pb.XUname {
	var out bytes.Buffer
	_, err := exec.LookPath("xuname")
	if err != nil {
		return nil
	}
	xunameCmd := exec.Command("xuname")
	xunameCmd.Stdout = &out
	if err := xunameCmd.Run(); err != nil {
		log.Fatal(err)
	}
	fields := strings.Fields(out.String())

	return &pb.XUname{
		OSName:       proto.String(fields[0]),
		Kernel:       proto.String(fields[1]),
		Mach:         proto.String(fields[2]),
		CPUInfo:      proto.String(fields[3]),
		UpdateStatus: proto.String(fields[4]),
		RepoStatus:   proto.String(fields[5]),
	}
}

func main() {
	flag.Parse()
	machineID = getUUID()

	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *server, *port), opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewPopCornClient(conn)

	q := pb.Stats{
		HostID: &machineID,
		Pkgs:   getPkgs(),
		XUname: getXUname(),
	}

	_, err = client.Update(context.Background(), &q)
	if err != nil {
		log.Fatal(err)
	}
}

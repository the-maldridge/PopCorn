package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/subcommands"
	"google.golang.org/grpc"

	pb "github.com/the-maldridge/popcorn/internal/proto"
)

var (
	server = flag.String("server", "localhost", "Server to connect to")
	port   = flag.Int("port", 8080, "Port to connect on")
)

// Get the report
type reportCmd struct{ path string }

func (*reportCmd) Name() string     { return "report" }
func (*reportCmd) Synopsis() string { return "Request a report from the server" }
func (*reportCmd) Usage() string {
	return `report --output <file>
Write a report to the specified file.
`
}
func (r *reportCmd) SetFlags(f *flag.FlagSet) {
	flag.StringVar(&r.path, "file", "output.json", "File to write the output to")
}
func (r *reportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *server, *port), opts...)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer conn.Close()

	client := pb.NewPopCornClient(conn)

	result, err := client.Report(context.Background(), &pb.ReportRequest{})
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	if err := ioutil.WriteFile(r.path, result.GetReport(), 0644); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func main() {
	flag.Parse()

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	subcommands.Register(&reportCmd{}, "Administration")

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}

package ctl

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/google/subcommands"
	"google.golang.org/grpc"

	pb "github.com/the-maldridge/popcorn/internal/proto"
)

// Get the report
type ReportCmd struct {
	path  string
	reset bool
	key   string
}

func (*ReportCmd) Name() string     { return "report" }
func (*ReportCmd) Synopsis() string { return "Request a report from the server" }
func (*ReportCmd) Usage() string {
	return `report --output <file> [--reset --key <key]
Write a report to the specified file.
`
}

func (r *ReportCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.path, "file", "output.json", "File to write the output to")
	f.BoolVar(&r.reset, "reset", false, "Reset the repository after getting the report")
	f.StringVar(&r.key, "key", "", "Key to reset the repo with")
}

func (r *ReportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", server, port), opts...)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer conn.Close()

	client := pb.NewPopCornClient(conn)

	req := pb.ReportRequest{
		ResetRepo: &r.reset,
		ResetKey:  &r.key,
	}

	result, err := client.Report(context.Background(), &req)
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

package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"

	"github.com/the-maldridge/popcorn/internal/ctl"
)

var (
	server = flag.String("server", "localhost", "Server to connect to")
	port   = flag.Int("port", 8080, "Port to connect on")
)

func main() {
	flag.Parse()

	ctl.SetServer(*server)
	ctl.SetPort(*port)

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	subcommands.Register(&ctl.ReportCmd{}, "Administration")

	subcommands.Register(&ctl.PkgQueryCmd{}, "PQuery")

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}

package ctl

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/subcommands"
	"github.com/hashicorp/go-version"
	"google.golang.org/grpc"

	pb "github.com/the-maldridge/popcorn/pkg/proto/pquery"
)

// Get the report
type PkgQueryCmd struct {
	pkg    string
	start  string
	end    string
	dur    time.Duration
	format string
}

func (*PkgQueryCmd) Name() string     { return "pkgstats" }
func (*PkgQueryCmd) Synopsis() string { return "Request stats on a particular package" }
func (*PkgQueryCmd) Usage() string {
	return `pkgstats \
  --pkg <pkgname> \
  [--start YYYY-MM-DD] \
  [--end YYYY-MM-DD] \
  [--interval interval] \
  [--format <format>]

Obtain a report over the specified time range.
`
}

func (r *PkgQueryCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.pkg, "pkg", "", "Package to query for")
	f.StringVar(&r.start, "start", "", "Start date in YYYY-MM-DD")
	f.StringVar(&r.end, "end", "", "End date in YYYY-MM-DD")
	f.DurationVar(&r.dur, "interval", time.Hour*24*30, "Duration to query for")
	f.StringVar(&r.format, "format", "quick", "Output format, one of 'quick'")
}

func (r *PkgQueryCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	// Parse the start and end times.
	intrvl := pb.DateRange{}

	if r.start != "" {
		start, err := time.Parse("2006-01-02", r.start)
		if err != nil {
			fmt.Printf("Parse error: %s", err)
			return subcommands.ExitFailure
		}
		t := start.Unix()
		intrvl.StartTime = &t
	}
	if r.end != "" {
		end, err := time.Parse("2006-01-02", r.end)
		if err != nil {
			fmt.Printf("Parse error: %s", err)
			return subcommands.ExitFailure
		}
		t := end.Unix()
		intrvl.EndTime = &t
	}

	intrvl.Duration = proto.String(r.dur.String())

	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", server, port), opts...)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer conn.Close()

	client := pb.NewPQueryClient(conn)

	req := pb.PkgStatQuery{
		PkgName:  &r.pkg,
		Interval: &intrvl,
	}

	result, err := client.GetPackageStats(context.Background(), &req)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	switch r.format {
	case "quick":
		printQuickStats(result)
	case "date":
		printDateStats(result)
	case "csv":
		printCSV(result)
	default:
		printQuickStats(result)
	}
	return subcommands.ExitSuccess
}

func printQuickStats(s *pb.PackageStats) {
	if len(s.GetVersions()) == 0 {
		fmt.Println("PopCorn doesn't know of any installs for this package")
		return
	}

	fmt.Printf("According to PopCorn data, there are at least %d installs on the following versions:\n", s.GetInstalls())

	versions := s.GetVersions()

	// Sort the versions into some sane arrangement
	sort.Slice(versions, func(i, j int) bool {
		left, err := version.NewVersion(strings.Replace(versions[i], "_", ".", 1))
		if err != nil {
			fmt.Println(err)
		}
		right, _ := version.NewVersion(strings.Replace(versions[j], "_", ".", 1))
		return left.LessThan(right)
	})

	for _, v := range versions {
		fmt.Printf("  %s\n", v)
	}

}

func printDateStats(s *pb.PackageStats) {
	stats := s.GetCalendarStats()

	sort.Slice(stats, func(i, j int) bool {
		left, err := time.Parse("2006-01-02", *stats[i].Date)
		if err != nil {
			fmt.Println(err)
		}
		right, err := time.Parse("2006-01-02", *stats[j].Date)
		if err != nil {
			fmt.Println(err)
		}
		return left.Before(right)
	})

	for _, s := range stats {
		fmt.Printf("%s | ", *s.Date)
		vers := s.GetVersions()
		sort.Slice(vers, func(i, j int) bool {
			left, err := version.NewVersion(strings.Replace(vers[i].GetVersion(), "_", ".", 1))
			if err != nil {
				fmt.Println(err)
			}
			right, _ := version.NewVersion(strings.Replace(vers[j].GetVersion(), "_", ".", 1))
			return left.LessThan(right)
		})

		for _, v := range vers {
			fmt.Printf("%s:%d ", v.GetVersion(), v.GetInstalls())
		}
		fmt.Printf("\n")
	}
}

func printCSV(s *pb.PackageStats) {
	seperator := ","

	// The versions need to be sorted for this to work
	versions := s.GetVersions()
	sort.Slice(versions, func(i, j int) bool {
		left, err := version.NewVersion(strings.Replace(versions[i], "_", ".", 1))
		if err != nil {
			fmt.Println(err)
		}
		right, _ := version.NewVersion(strings.Replace(versions[j], "_", ".", 1))
		return left.LessThan(right)
	})

	// Print out the version headers
	fmt.Printf("Date%s", seperator)
	for i, v := range versions {
		fmt.Printf("%s", v)
		i++
		if i >= len(versions) {
			fmt.Printf("\n")
		} else {
			fmt.Printf(seperator)
		}
	}

	// Print the data
	// The data needs to be sorted by date
	stats := s.GetCalendarStats()
	sort.Slice(stats, func(i, j int) bool {
		left, err := time.Parse("2006-01-02", *stats[i].Date)
		if err != nil {
			fmt.Println(err)
		}
		right, err := time.Parse("2006-01-02", *stats[j].Date)
		if err != nil {
			fmt.Println(err)
		}
		return left.Before(right)
	})

	for _, s := range stats {
		fmt.Printf("%s%s", *s.Date, seperator)
		vers := s.GetVersions()
		sort.Slice(vers, func(i, j int) bool {
			left, err := version.NewVersion(strings.Replace(vers[i].GetVersion(), "_", ".", 1))
			if err != nil {
				fmt.Println(err)
			}
			right, _ := version.NewVersion(strings.Replace(vers[j].GetVersion(), "_", ".", 1))
			return left.LessThan(right)
		})

		// Print out the versions
		for i, j := 0, 0; i < len(versions); i++ {
			if versions[i] == vers[j].GetVersion() {
				fmt.Printf("%d", vers[j].GetInstalls())
				if j+1 <= len(vers)-1 {
					j++
				}
			} else {
				fmt.Printf("0")
			}
			if i != len(versions)-1 {
				fmt.Printf(seperator)
			}
		}
		fmt.Printf("\n")
	}
}

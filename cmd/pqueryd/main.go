package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	"github.com/the-maldridge/popcorn/internal/repo"

	pqpb "github.com/the-maldridge/popcorn/pkg/proto/pquery"
)

var (
	data_dir = flag.String("data_dir", "./data", "Directory containing PopCorn stats files")
	addr     = flag.String("addr", "", "Address to bind on")
	port     = flag.Int("port", 8081, "Port to bind on")

	statmap = make(map[time.Time]*repo.StatsRepo)
)

type PQueryServer struct{}

func parseInterval(dr *pqpb.DateRange) (time.Time, time.Time) {
	endTime := time.Now()
	startTime := endTime.Add(time.Hour * 24 * -30)
	if dr.GetEndTime() != 0 {
		endTime = time.Unix(dr.GetEndTime(), 0)
	}
	if dr.GetDuration() != "" {
		d, err := time.ParseDuration(dr.GetDuration())
		if err != nil {
			d = time.Hour * 24 * 30
		}
		startTime = endTime.Add(d * -1)
	}
	if dr.GetStartTime() != 0 {
		startTime = time.Unix(dr.GetStartTime(), 0)
	}
	return startTime, endTime
}

func (*PQueryServer) GetPackageStats(ctx context.Context, r *pqpb.PkgStatQuery) (*pqpb.PackageStats, error) {
	// Parse out the time interval
	startTime, endTime := parseInterval(r.GetInterval())

	// Scrape the data out of the statmap
	pkgname := r.GetPkgName()
	pkgdays := []*pqpb.PkgDay{}

	knownInstalls := 0
	knownVersions := make(map[string]bool)

	// Iterate on days
	for day, stats := range statmap {
		// Check if day is in range
		if day.After(startTime) && day.Before(endTime) {
			// Check if this package is present
			if _, ok := stats.Versions[pkgname]; ok {
				// Construct the versions
				versions := []*pqpb.PkgVersion{}
				i := 0
				for v, c := range stats.Versions[pkgname] {
					knownVersions[v] = true
					i += c
					versionsForDay := pqpb.PkgVersion{
						Version:  proto.String(v),
						Installs: proto.Int32(int32(c)),
					}
					versions = append(versions, &versionsForDay)
				}
				if i > knownInstalls {
					knownInstalls = i
				}
				dStr := day.Format("2006-01-02")
				day := pqpb.PkgDay{
					Date:     &dStr,
					Versions: versions,
				}
				pkgdays = append(pkgdays, &day)
			}
		}
	}

	// Dump versions to a list of strings
	vers := []string{}
	for v, _ := range knownVersions {
		vers = append(vers, v)
	}

	// Compose the reply
	stats := pqpb.PackageStats{
		Installs:      proto.Int32(int32(knownInstalls)),
		Versions:      vers,
		CalendarStats: pkgdays,
	}
	// Return the stats data
	return &stats, nil
}

func loadData() {
	globs, err := filepath.Glob(filepath.Join(*data_dir, "popcorn_*.json"))
	if err != nil {
		log.Fatalf("Error loading data: %s", err)
	}

	log.Printf("Loaded %d stats files", len(globs))

	for _, g := range globs {
		d, err := ioutil.ReadFile(g)
		if err != nil {
			log.Printf("Error loading %s: %s", g, err)
			continue
		}

		dateStr := strings.Replace(filepath.Base(g), "popcorn_", "", -1)
		dateStr = strings.Replace(dateStr, ".json", "", -1)

		statset := &repo.StatsRepo{}
		err = json.Unmarshal(d, statset)
		if err != nil {
			log.Printf("Error loading %s: %s", g, err)
			continue
		}
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Printf("Date parsing error: %s", err)
			continue
		}
		statmap[date] = statset
	}
}

func main() {
	flag.Parse()

	log.Println("pqueryd is starting...")
	log.Printf("Stats files will be read from %s", *data_dir)

	loadData()

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *addr, *port))
	if err != nil {
		log.Fatal(err)
	}

	var opts []grpc.ServerOption
	srvr := grpc.NewServer(opts...)
	pqpb.RegisterPQueryServer(srvr, &PQueryServer{})
	srvr.Serve(l)
}

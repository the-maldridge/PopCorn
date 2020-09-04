package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/the-maldridge/popcorn/pkg/stats"

	pb "github.com/the-maldridge/popcorn/internal/proto"
)

var (
	backoff = []time.Duration{
		5 * time.Second,
		5 * time.Second,
		5 * time.Second,
		10 * time.Second,
		10 * time.Second,
		10 * time.Second,
		1 * time.Minute,
		1 * time.Minute,
		1 * time.Minute,
	}
)

type StatsRepo struct {
	c *http.Client
}

func New() *StatsRepo {
	return &StatsRepo{
		c: &http.Client{
			Timeout: time.Second * 5,
		},
	}
}

func (s *StatsRepo) Update(ctx context.Context, r *pb.Stats) (*pb.StatsConfirmation, error) {
	log.Printf("Proxy request from %s", r.GetHostID())

	pkgs := []stats.Package{}
	for _, p := range r.GetPkgs() {
		pkgs = append(pkgs, stats.Package{
			Name:    p.GetName(),
			Version: p.GetVersion(),
		})
	}

	xn := stats.XUname{}
	xn.OSName = r.GetXUname().GetOSName()
	xn.Kernel = r.GetXUname().GetKernel()
	xn.Mach = r.GetXUname().GetMach()
	xn.CPUInfo = r.GetXUname().GetCPUInfo()
	xn.UpdateStatus = r.GetXUname().GetUpdateStatus()
	xn.RepoStatus = r.GetXUname().GetRepoStatus()

	d := stats.Stats{
		Packages: pkgs,
		XUname:   xn,
	}

	body, err := json.Marshal(d)
	if err != nil {
		return &pb.StatsConfirmation{}, err
	}

	req, err := http.NewRequest(http.MethodPost, os.Getenv("STATS_URL"), bytes.NewBuffer(body))
	if err != nil {
		return &pb.StatsConfirmation{}, err
	}
	req.Header.Add("From", r.GetHostID())
	req.Header.Add("Content-type", "application/json")

	for _, td := range backoff {
		_, err := s.c.Do(req)
		if err != nil {
			time.Sleep(td)
			continue
		}
		return &pb.StatsConfirmation{}, nil
	}

	return &pb.StatsConfirmation{}, nil
}

func (s *StatsRepo) Report(ctx context.Context, r *pb.ReportRequest) (*pb.StatsReport, error) {
	return &pb.StatsReport{}, nil
}

func (s *StatsRepo) Ping(ctx context.Context, r *pb.EmptyRequest) (*pb.EmptyResponse, error) {
	return &pb.EmptyResponse{}, nil
}

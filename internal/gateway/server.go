package gateway

import (
	"context"
	"flag"
	"log"

	"github.com/the-maldridge/popcorn/internal/repo"
	"google.golang.org/grpc/peer"

	pb "github.com/the-maldridge/popcorn/internal/proto"
)

var (
	resetKey = flag.String("reset_key", "", "Key required to reset the stats repo")
)

type StatsRepo struct {
	repo *repo.StatsRepo
}

func New(r *repo.StatsRepo) *StatsRepo {
	return &StatsRepo{
		repo: r,
	}
}

func (s *StatsRepo) Update(ctx context.Context, r *pb.Stats) (*pb.StatsConfirmation, error) {
	s.repo.AddStats(*r)
	p, ok := peer.FromContext(ctx)
	if ok {
		log.Printf("Stat update from %s", p.Addr)
	} else {
		log.Println("Stat update from unidentified peer")
	}

	return &pb.StatsConfirmation{}, nil
}

func (s *StatsRepo) Report(ctx context.Context, r *pb.ReportRequest) (*pb.StatsReport, error) {
	d, err := s.repo.GetReport()
	if err != nil {
		return nil, err
	}
	log.Println("Report Requested")

	if r.GetResetRepo() && r.GetResetKey() == *resetKey {
		s.repo.Reset()
		log.Println("Stats repo reset")
	}

	return &pb.StatsReport{Report: d}, nil
}

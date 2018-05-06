package gateway

import (
	"context"
	"log"

	"google.golang.org/grpc/peer"
	"github.com/the-maldridge/popcorn/internal/repo"

	pb "github.com/the-maldridge/popcorn/internal/proto"
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
	return &pb.StatsReport{Report: d}, nil
}

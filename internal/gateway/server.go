package gateway

import (
	"context"
	"log"

	pb "github.com/the-maldridge/popcorn/internal/proto"
)

type StatsRepo struct{}

func (s *StatsRepo) Update(ctx context.Context, r *pb.Stats) (*pb.StatsConfirmation, error) {
	log.Printf("Obtained stats: %v", r)

	return &pb.StatsConfirmation{}, nil
}

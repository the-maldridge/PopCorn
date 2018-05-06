package gateway

import (
	"context"

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

	return &pb.StatsConfirmation{}, nil
}

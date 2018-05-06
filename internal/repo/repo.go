package repo

import (
	"log"
	"time"

	pb "github.com/the-maldridge/popcorn/internal/proto"
)

type StatsRepo struct {
	lastSeen       map[string]time.Time
	packages       map[string]int
	versions       map[string]map[string]int
	xuOSName       map[string]int
	xuKernel       map[string]int
	xuMach         map[string]int
	xuCPUInfo      map[string]int
	xuUpdateStatus map[string]int
	xuRepoStatus   map[string]int
}

func New() *StatsRepo {
	return &StatsRepo{
		lastSeen:       make(map[string]time.Time),
		packages:       make(map[string]int),
		versions:       make(map[string]map[string]int),
		xuOSName:       make(map[string]int),
		xuKernel:       make(map[string]int),
		xuMach:         make(map[string]int),
		xuCPUInfo:      make(map[string]int),
		xuUpdateStatus: make(map[string]int),
		xuRepoStatus:   make(map[string]int),
	}
}

func (r *StatsRepo) AddStats(s pb.Stats) {
	r.lastSeen[s.GetHostID()] = time.Now()

	for _, p := range s.GetPkgs() {
		r.packages[p.GetName()]++
		if r.versions[p.GetName()] == nil {
			r.versions[p.GetName()] = make(map[string]int)
		}
		r.versions[p.GetName()][p.GetVersion()]++
	}

	if s.GetXUname() != nil {
		x := s.GetXUname()
		r.xuOSName[x.GetOSName()]++
		r.xuKernel[x.GetKernel()]++
		r.xuMach[x.GetMach()]++
		r.xuCPUInfo[x.GetCPUInfo()]++
		r.xuUpdateStatus[x.GetUpdateStatus()]++
		r.xuRepoStatus[x.GetRepoStatus()]++
	}

	log.Println(r)
}

package repo

import (
	"log"
	"time"

	pb "github.com/the-maldridge/popcorn/internal/proto"
)

type StatsRepo struct {
	lastSeen map[string]time.Time
	packages map[string]int
	versions map[string]map[string]int
	xuname   map[string]map[string]int
}

func New() *StatsRepo {
	return &StatsRepo{
		lastSeen: make(map[string]time.Time),
		packages: make(map[string]int),
		versions: make(map[string]map[string]int),
		xuname:   make(map[string]map[string]int),
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
		r.xuname["OSName"][x.GetOSName()]++
		r.xuname["Kernel"][x.GetKernel()]++
		r.xuname["Mach"][x.GetMach()]++
		r.xuname["CPUInfo"][x.GetCPUInfo()]++
		r.xuname["UpdateStatus"][x.GetUpdateStatus()]++
		r.xuname["RepoStatus"][x.GetRepoStatus()]++
	}

	log.Println(r)
}

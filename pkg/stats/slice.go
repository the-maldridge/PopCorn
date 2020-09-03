package stats

import (
	"strings"
)

// NewRDS returns a new RepoDataSlice
func NewRDS() *RepoDataSlice {
	r := &RepoDataSlice{}

	r.Seen = make(map[string]struct{})
	r.Packages = make(map[string]int)
	r.Versions = make(map[string]map[string]int)
	r.XuOSName = make(map[string]int)
	r.XuKernel = make(map[string]int)
	r.XuMach = make(map[string]int)
	r.XuCPUInfo = make(map[string]int)
	r.XuUpdateStatus = make(map[string]int)
	r.XuRepoStatus = make(map[string]int)

	return r
}

// AddStats adds some new stats if the ID is unique and not yet seen.
func (rds *RepoDataSlice) AddStats(id string, s Stats) {
	id = strings.TrimSpace(id)

	if _, ok := rds.Seen[id]; ok {
		return
	}

	rds.Seen[id] = struct{}{}
	rds.UniqueInstalls++

	for i := range s.Packages {
		rds.Packages[s.Packages[i].Name]++
		if rds.Versions[s.Packages[i].Name] == nil {
			rds.Versions[s.Packages[i].Name] = make(map[string]int)
		}
		rds.Versions[s.Packages[i].Name][s.Packages[i].Version]++
	}

	if s.XUname.OSName != "" {
		rds.XuOSName[s.XUname.OSName]++
		rds.XuKernel[s.XUname.Kernel]++
		rds.XuMach[s.XUname.Mach]++
		rds.XuCPUInfo[s.XUname.CPUInfo]++
		rds.XuUpdateStatus[s.XUname.UpdateStatus]++
		rds.XuRepoStatus[s.XUname.RepoStatus]++
	}
}
package repo

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	pb "github.com/the-maldridge/popcorn/internal/proto"
)

var (
	checkpointFile     = flag.String("checkpoint_file", "/var/lib/popcorn/checkpoint.json", "Location of checkpoint file")
	checkpointInterval = flag.Duration("checkpoint_interval", 15*time.Minute, "Frequency of checkpoints")

	needCheckpoint = false
)

type StatsRepo struct {
	UniqueInstalls int
	LastSeen       map[string]time.Time
	Packages       map[string]int
	Versions       map[string]map[string]int
	XuOSName       map[string]int
	XuKernel       map[string]int
	XuMach         map[string]int
	XuCPUInfo      map[string]int
	XuUpdateStatus map[string]int
	XuRepoStatus   map[string]int
}

func New() *StatsRepo {
	r := &StatsRepo{
		LastSeen:       make(map[string]time.Time),
		Packages:       make(map[string]int),
		Versions:       make(map[string]map[string]int),
		XuOSName:       make(map[string]int),
		XuKernel:       make(map[string]int),
		XuMach:         make(map[string]int),
		XuCPUInfo:      make(map[string]int),
		XuUpdateStatus: make(map[string]int),
		XuRepoStatus:   make(map[string]int),
	}

	go r.checkpointTimer()

	d, err := ioutil.ReadFile(*checkpointFile)
	if os.IsNotExist(err) {
		// No checkpoint
		return r
	}
	if err := json.Unmarshal(d, r); err != nil {
		log.Println("Checkpoint Reload Error")
		return r
	}
	log.Println("Checkpoint Reloaded")
	return r
}

func (r *StatsRepo) AddStats(s pb.Stats) {
	r.LastSeen[s.GetHostID()] = time.Now()
	r.UniqueInstalls = len(r.LastSeen)

	for _, p := range s.GetPkgs() {
		r.Packages[p.GetName()]++
		if r.Versions[p.GetName()] == nil {
			r.Versions[p.GetName()] = make(map[string]int)
		}
		r.Versions[p.GetName()][p.GetVersion()]++
	}

	if s.GetXUname() != nil {
		x := s.GetXUname()
		r.XuOSName[x.GetOSName()]++
		r.XuKernel[x.GetKernel()]++
		r.XuMach[x.GetMach()]++
		r.XuCPUInfo[x.GetCPUInfo()]++
		r.XuUpdateStatus[x.GetUpdateStatus()]++
		r.XuRepoStatus[x.GetRepoStatus()]++
	}

	needCheckpoint = true
}

func (r *StatsRepo) GetReport() ([]byte, error) {
	d, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *StatsRepo) checkpoint() {
	d, err := r.GetReport()
	if err != nil {
		log.Println("Error checkpointing")
		return
	}

	if err := ioutil.WriteFile(*checkpointFile, d, 0644); err != nil {
		log.Println("Error writing checkpoint file")
		return
	}
	log.Println("State checkpointed")
}

func (r *StatsRepo) checkpointTimer() {
	for range time.Tick(*checkpointInterval) {
		if needCheckpoint {
			r.checkpoint()
			needCheckpoint = false
		}
	}
}

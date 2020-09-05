package stats

import (
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
)

// A Store is a mechanism that can persist stats in a long term
// fashion.
type Store interface {
	PutSlice(string, *RepoDataSlice) error
	GetSlice(string) (*RepoDataSlice, error)

	ListSlices() ([]string, error)
}

// A Repo has a set of methods for accepting stats and for then
// persisting those aggregate stats to a local or remote store.
type Repo struct {
	*echo.Echo

	store Store

	cron *cron.Cron

	log          hclog.Logger
	currentSlice *RepoDataSlice
	currentKey   string
}

// A RepoDataSlice is the active slice that a repo server is acting on
// at any given time.
type RepoDataSlice struct {
	mutex sync.RWMutex

	dirty bool

	UniqueInstalls int
	Seen           map[string]struct{}
	Packages       map[string]int
	Versions       map[string]map[string]int
	XuOSName       map[string]int
	XuKernel       map[string]int
	XuMach         map[string]int
	XuCPUInfo      map[string]int
	XuUpdateStatus map[string]int
	XuRepoStatus   map[string]int
}

// Stats composes a stats package from a host.
type Stats struct {
	Packages []Package
	XUname   XUname
}

// A Package is a name and version as reported by xbps-query.
type Package struct {
	Name    string
	Version string
}

// XUname structures the output as reported by xuname from xtools.
type XUname struct {
	OSName       string
	Kernel       string
	Mach         string
	CPUInfo      string
	UpdateStatus string
	RepoStatus   string
}

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

// A StoreFactory returns a store who's logger is parented to the
// logger provided to the factory.
type StoreFactory func(hclog.Logger) (Store, error)

// A Callback is a function that can be used to register a
// StoreFactory to the list of stores that are available.  This allows
// certain initialization tasks to be deferred until after config file
// loading and logger setup have completed.
type Callback func()

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

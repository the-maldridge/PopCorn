package stats

import (
	"errors"

	"github.com/hashicorp/go-hclog"
)

var (
	factories map[string]StoreFactory

	storeParentLogger hclog.Logger
)

func init() {
	factories = make(map[string]StoreFactory)
}

// RegisterStore is called to register a store implementation to the
// repo.
func RegisterStore(n string, f StoreFactory) {
	if _, exists := factories[n]; exists {
		return
	}
	log().Info("Registered store", "store", n)
	factories[n] = f
}

// NewStore attempts to construct a store instance and return it.
func NewStore(n string) (Store, error) {
	f, exists := factories[n]
	if !exists {
		return nil, errors.New("no factory exists with that name")
	}
	return f(log())
}

// SetStoreParentLogger plumbs in the parent logger for package level
// logging needs.
func SetStoreParentLogger(l hclog.Logger) {
	storeParentLogger = l.Named("storage")
}

func log() hclog.Logger {
	if storeParentLogger == nil {
		return hclog.NewNullLogger()
	}
	return storeParentLogger
}

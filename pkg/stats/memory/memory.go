package memory

import (
	"errors"
	"sync"

	"github.com/the-maldridge/popcorn/pkg/stats"
)

type mem struct {
	mutex sync.RWMutex

	d map[string]*stats.RepoDataSlice
}

// New returns an in memory implementation of stats.Store from the
// stats package.
func New() stats.Store {
	m := new(mem)
	m.d = make(map[string]*stats.RepoDataSlice)
	return m
}

func (m *mem) PutSlice(k string, v *stats.RepoDataSlice) error {
	m.mutex.Lock()
	m.d[k] = v
	m.mutex.Unlock()
	return nil
}

func (m *mem) GetSlice(k string) (*stats.RepoDataSlice, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	s, ok := m.d[k]
	if !ok {
		return nil, errors.New("slice not found")
	}
	return s, nil
}

package memory

import (
	"errors"
	"sync"

	"github.com/hashicorp/go-hclog"

	"github.com/the-maldridge/popcorn/pkg/stats"
)

type mem struct {
	mutex sync.RWMutex

	d map[string]*stats.RepoDataSlice
}

func init() {
	stats.RegisterCallback(cb)
}

func cb() {
	stats.RegisterStore("memory", New)
}

// New returns an in memory implementation of stats.Store from the
// stats package.
func New(hclog.Logger) (stats.Store, error) {
	m := new(mem)
	m.d = make(map[string]*stats.RepoDataSlice)
	return m, nil
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

func (m *mem) ListSlices() ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	o := make([]string, len(m.d))
	i := 0
	for k := range m.d {
		o[i] = k
		i++
	}
	return o, nil
}

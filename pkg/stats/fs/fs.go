package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"

	"github.com/the-maldridge/popcorn/pkg/stats"
)

type fs struct {
	basePath string
	prefix   string
}

func init() {
	stats.RegisterCallback(cb)
}

func cb() {
	stats.RegisterStore("filesystem", New)
}

// New constructs a filesystem store.
func New(l hclog.Logger) (stats.Store, error) {
	x := new(fs)

	x.basePath = os.Getenv("FS_BASE_PATH")
	x.prefix = os.Getenv("FS_PREFIX")

	if x.prefix == "" {
		x.prefix = "popcorn_"
	}

	return x, nil
}

func (fs *fs) PutSlice(k string, s *stats.RepoDataSlice) error {
	bytes, err := json.Marshal(s)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(fs.basePath, fs.prefix+k+".json"), bytes, 0644); err != nil {
		return err
	}

	return nil
}

func (fs *fs) GetSlice(k string) (*stats.RepoDataSlice, error) {
	in, err := ioutil.ReadFile(filepath.Join(fs.basePath, fs.prefix+k+".json"))
	switch {
	case os.IsNotExist(err):
		return nil, stats.ErrNoSuchSlice
	case err != nil:
		return nil, err
	}

	slice := stats.NewRDS()
	if err := json.Unmarshal(in, &slice); err != nil {
		return nil, err
	}
	return slice, nil
}

func (fs *fs) ListSlices() ([]string, error) {
	globs, _ := filepath.Glob(filepath.Join(fs.basePath, fs.prefix+"*.json"))
	keys := make([]string, len(globs))
	for i := range globs {
		t := globs[i]
		t = filepath.Base(t)
		t = strings.Replace(t, fs.prefix, "", 1)
		t = strings.Replace(t, ".json", "", 1)
		keys[i] = t
	}
	return keys, nil
}

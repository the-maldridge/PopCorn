package stats

import (
	"context"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
)

const (
	keyfmt = "2006-01-02"

	syncCron = "*/15 * * * *"
	rotCron = "1 0 * * *"
)

// New returns an repo server that can accept stats and produce
// reports.
func New(l hclog.Logger, s Store) *Repo {
	r := &Repo{}
	r.log = l.Named("http")
	r.store = s

	r.Echo = echo.New()

	p := prometheus.NewPrometheus("echo", nil)
	p.Use(r.Echo)

	r.POST("/v1/stats/add", r.addStats)
	r.GET("/v1/stats/:key", r.getStats)

	r.currentKey = time.Now().Format(keyfmt)
	r.currentSlice = r.loadSlice(r.currentKey)

	r.cron = cron.New()
	r.cron.AddFunc(syncCron, r.sync)
	r.cron.AddFunc(rotCron, r.rotate)

	return r
}

// Serve serves the repo via HTTP on the given bind string
func (r *Repo) Serve(b string) error {
	r.cron.Start()
	return r.Start(b)
}

// Shutdown calls all deferred shutdown tasks before releasing the
// server instance.
func (r *Repo) Shutdown(ctx context.Context) {
	r.log.Debug("Stopping cron routines")
	r.cron.Stop()
	r.log.Debug("Shutting down webserver")
	r.Echo.Shutdown(ctx)
	r.log.Debug("Performing final slice sync")
	r.sync()
}

// sync flushes the current slice to disk.
func (r *Repo) sync() {
	if !r.currentSlice.dirty {
		return
	}

	if err := r.store.PutSlice(r.currentKey, r.currentSlice); err != nil {
		r.log.Error("Error persisting slice data", "error", err)
	}
}

// loadSlice tries to load the current slice or will create a new
// slice ot use if the current slice cannot be found.
func (r *Repo) loadSlice(k string) *RepoDataSlice {
	s, err := r.store.GetSlice(k)
	if err != nil {
		r.log.Warn("No slice for key, returning a new one", "key", k)
		return NewRDS()
	}
	return s
}

// rotate handles the daily changeover of rotating and persisting
// slice data.  Importantly it does this in a semi-atomic fashion
// where it shouldn't be possible to lose stats after the slice
// pointer has been rotated.
func (r *Repo) rotate() {
	r.log.Debug("Rotating data slice")
	oSlice := r.currentSlice
	oKey := r.currentKey

	r.currentKey = time.Now().Format(keyfmt)
	r.currentSlice = r.loadSlice(r.currentKey)

	if err := r.store.PutSlice(oKey, oSlice); err != nil {
		r.log.Error("Error persisting slice", "error", err)
		return
	}
	r.log.Info("Data slice rotated", "current", r.currentKey, "old", oKey)
}

package stats

import (
	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
)

// New returns an repo server that can accept stats and produce
// reports.
func New(l hclog.Logger) *Repo {
	r := &Repo{}
	r.Echo = echo.New()
	r.log = l.Named("http")

	p := prometheus.NewPrometheus("echo", nil)
	p.Use(r.Echo)

	r.POST("/v1/stats/add", r.addStats)
	r.GET("/v1/stats/current", r.getCurrentStats)

	r.currentSlice = NewRDS()

	return r
}

// Serve serves the repo via HTTP on the given bind string
func (r *Repo) Serve(b string) error {
	return r.Start(b)
}

package stats

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (r *Repo) addStats(c echo.Context) error {
	f, ok := c.Request().Header["From"]
	if !ok {
		return c.JSON(
			http.StatusPreconditionFailed,
			struct{ Message string }{Message: "From header is required"},
		)
	}

	s := Stats{}
	if err := c.Bind(&s); err != nil {
		return err
	}

	r.currentSlice.AddStats(f[0], s)
	return c.NoContent(http.StatusOK)
}

func (r *Repo) getStats(c echo.Context) error {
	if c.Param("key") == "current" {
		return c.JSON(http.StatusOK, r.currentSlice)
	}

	return c.JSON(http.StatusOK, r.loadSlice(c.Param("key")))
}

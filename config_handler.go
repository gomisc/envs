package envs

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"gopkg.in/gomisc/errors.v1"
)

const (
	apiPrefix   = "/api"
	prefixParam = "prefix"
	keyParam    = "key"
	valParam    = "val"
	delimParam  = "delim"
)

func (ctl *localConfigController) router() *echo.Echo {
	router := echo.New()
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())

	router.GET(
		"/alive", func(ctx echo.Context) error {
			return ctx.String(http.StatusOK, "alive")
		},
	)

	api := router.Group("/api")
	api.GET("/:"+keyParam, ctl.get)
	api.GET("/:"+prefixParam+"/:"+keyParam, ctl.getFor)
	api.PUT("/:"+keyParam+"/:"+valParam, ctl.set)
	api.PUT("/:"+prefixParam+"/:"+keyParam+"/:"+valParam, ctl.setFor)
	api.POST("/:"+keyParam+"/:"+valParam, ctl.add)
	api.POST("/:"+prefixParam+"/:"+keyParam+"/:"+valParam, ctl.addFor)
	api.GET("/dump", ctl.dump)
	api.GET("/:"+prefixParam+"/dump", ctl.dumpFor)

	return router
}

// PUT http://host:port/api/key/value -> 200 OK
func (ctl *localConfigController) set(c echo.Context) error {
	ctl.Set(c.Param(keyParam), c.Param(valParam))

	return c.String(http.StatusOK, "ok")
}

// PUT http://host:port/api/prefix/key/value -> 200 OK
func (ctl *localConfigController) setFor(c echo.Context) error {
	ctl.SetFor(c.Param(prefixParam), c.Param(keyParam), c.Param(valParam))

	return c.String(http.StatusOK, "ok")
}

// POST http://host:port/api/key/value?delim=delim -> 200 OK
func (ctl *localConfigController) add(c echo.Context) error {
	ctl.Add(c.Param(keyParam), c.Param(valParam), c.QueryParam(delimParam))

	return c.String(http.StatusOK, "ok")
}

// POST http://host:port/api/prefix/key/value?delim=delim -> 200 OK
func (ctl *localConfigController) addFor(c echo.Context) error {
	ctl.AddFor(c.Param(prefixParam), c.Param(keyParam), c.Param(valParam), c.QueryParam(delimParam))

	return c.String(http.StatusOK, "ok")
}

// GET http://host:port/api/key -> 200 OK "value"
func (ctl *localConfigController) get(c echo.Context) error {
	val, ok := ctl.Get(c.Param(keyParam))

	if ok {
		return c.String(http.StatusOK, val)
	}

	return c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

// GET http://host:port/api/prefix/key -> 200 OK "value"
func (ctl *localConfigController) getFor(c echo.Context) error {
	val, ok := ctl.GetFor(c.Param(prefixParam), c.Param(keyParam))

	if ok {
		return c.String(http.StatusOK, val)
	}

	return c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

// GET http://host:port/api -> 200 OK [string...]
// empty | ["nameVar1", "nameVar2", ..."nameVarN"]
// если не заданы параметры фильтра -  вернет все пары ключ=значение
func (ctl *localConfigController) dump(c echo.Context) error {
	var filter []string

	if err := c.Bind(&filter); err != nil {
		return errors.Wrap(err, "parse filter")
	}

	dump := ctl.DumpEnv(filter...)

	return c.JSONPretty(http.StatusOK, &dump, "  ")
}

// GET http://host:port/api -> 200 OK [string...]
// empty | ["nameVar1", "nameVar2", ..."nameVarN"]
// если не заданы параметры фильтра -  вернет все пары ключ=значение
func (ctl *localConfigController) dumpFor(c echo.Context) error {
	var filter []string

	if err := c.Bind(&filter); err != nil {
		return errors.Wrap(err, "parse filter")
	}

	dump := ctl.DumpEnvFor(c.Param(prefixParam), filter...)

	return c.JSONPretty(http.StatusOK, &dump, "  ")
}

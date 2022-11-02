package controllers

import (
	"net/http"
	"strings"

	"git.eth4.dev/golibs/network/http"
	"git.eth4.dev/golibs/slog"
)

type remoteConfigController struct {
	log      slog.Logger
	endpoint string
}

// RemoteConfigController - конструктор контроллера конфигурации с подключением
// через REST интерфейс
func RemoteConfigController(log slog.Logger, host string) Controller {
	return &remoteConfigController{
		log:      log,
		endpoint: host + apiPrefix,
	}
}

// Endpoint - точка подключения к контроллеру
func (c *remoteConfigController) Endpoint() string {
	if c == nil {
		return ""
	}

	return c.endpoint
}

// Set PUT http://host:port/api/key/value -> 200 OK
func (c *remoteConfigController) Set(key, value string) {
	if c == nil {
		return
	}

	resp, err := nethttp.Request(
		strings.Join([]string{c.endpoint, key, value}, "/"),
	).Put()
	if err != nil {
		c.log.Error("send set request", err)
	}

	if err = nethttp.ResponseOrError(resp, http.StatusOK, nil); err != nil {
		c.log.Error("check response", err)
	}
}

// SetFor PUT http://host:port/api/prefix/key/value -> 200 OK
func (c *remoteConfigController) SetFor(prefix, key, value string) {
	if c == nil {
		return
	}

	resp, err := nethttp.Request(
		strings.Join([]string{c.endpoint, prefix, key, value}, "/"),
	).Put()
	if err != nil {
		c.log.Error("send set request", err)
	}

	if err = nethttp.ResponseOrError(resp, http.StatusOK, nil); err != nil {
		c.log.Error("check response", err)
	}
}

// Add - POST http://host:port/api/key/value?delim=delim -> 200 OK
func (c *remoteConfigController) Add(key, value, delim string) {
	if c == nil {
		return
	}

	resp, err := nethttp.Request(
		strings.Join([]string{c.endpoint, key, value}, "/") + "?delim=" + delim,
	).Post()
	if err != nil {
		c.log.Error("send request", err)

		return
	}

	if err = nethttp.ResponseOrError(resp, http.StatusOK, nil); err != nil {
		c.log.Error("check response", err)
	}
}

// AddFor - POST http://host:port/api/prefix/key/value?delim=delim -> 200 OK
func (c *remoteConfigController) AddFor(prefix, key, value, delim string) {
	if c == nil {
		return
	}

	resp, err := nethttp.Request(
		strings.Join([]string{c.endpoint, prefix, key, value}, "/") + "?delim=" + delim,
	).Post()
	if err != nil {
		c.log.Error("send request", err)

		return
	}

	if err = nethttp.ResponseOrError(resp, http.StatusOK, nil); err != nil {
		c.log.Error("check response", err)
	}
}

// Get GET http://host:port/api/key -> 200 OK string
func (c *remoteConfigController) Get(key string) (value string, ok bool) {
	if c == nil {
		return "", false
	}

	resp, err := nethttp.Request(strings.Join([]string{c.endpoint, key}, "/")).Get()
	if err != nil {
		c.log.Error("send request", err)

		return "", false
	}

	if err = nethttp.ResponseOrError(resp, http.StatusOK, &value); err != nil {
		c.log.Error("check response", err)

		return "", false
	}

	return value, true
}

// GetFor GET http://host:port/api/prefix/key -> 200 OK string
func (c *remoteConfigController) GetFor(prefix, key string) (value string, ok bool) {
	if c == nil {
		return "", false
	}

	resp, err := nethttp.Request(strings.Join([]string{c.endpoint, prefix, key}, "/")).Get()
	if err != nil {
		c.log.Error("send request", err)

		return "", false
	}

	if err = nethttp.ResponseOrError(resp, http.StatusOK, &value); err != nil {
		c.log.Error("check response", err)

		return "", false
	}

	return value, true
}

// DumpEnv GET http://host:port/api/dump -> 200 OK []string
func (c *remoteConfigController) DumpEnv(filter ...string) []string {
	if c == nil {
		return nil
	}

	resp, err := nethttp.Request(c.endpoint+"/dump").
		Header(nethttp.HeaderContentType, "application/json").
		Body(filter).
		Get()
	if err != nil {
		c.log.Error("send request", err)

		return nil
	}

	dump := []string{}
	if err = nethttp.ResponseOrError(resp, http.StatusOK, &dump); err != nil {
		c.log.Error("check response", err)

		return nil
	}

	return dump
}

// DumpEnvFor GET http://host:port/api/dump -> 200 OK []string
func (c *remoteConfigController) DumpEnvFor(prefix string, filter ...string) []string {
	if c == nil {
		return nil
	}

	resp, err := nethttp.Request(strings.Join([]string{c.endpoint, prefix, "dump"}, "/")).
		Header(nethttp.HeaderContentType, "application/json").
		Body(filter).
		Get()
	if err != nil {
		c.log.Error("send request", err)

		return nil
	}

	dump := []string{}
	if err = nethttp.ResponseOrError(resp, http.StatusOK, &dump); err != nil {
		c.log.Error("check response", err)

		return nil
	}

	return dump
}

// Close имплементтация io.Closer
func (c *remoteConfigController) Close() error {
	return nil
}

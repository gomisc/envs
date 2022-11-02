package controllers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"git.eth4.dev/golibs/errors"
	"git.eth4.dev/golibs/slog"
)

const (
	// ConfigControllerPortKey - ключ для хранения порта контроллера конфигурации
	ConfigControllerPortKey = "/confctl/port"
)

// Controller - контроллер конфигурирования тестовой среды
type Controller interface {
	// Endpoint возвращает адрес сервера контроллера
	Endpoint() string
	// Set устанавливает значение переменной окружения конфига
	Set(key, value string)
	// SetFor устанавливает опцию конфига для префикса
	SetFor(prefix, key, value string)
	// Add - устанавливает, либо добавляет значение в существующую переменную окружения конфига через указанный разделитель
	Add(key, value, delim string)
	// AddFor устанавливает опцию конфига для префикса, либо добавляет значение
	// в существующую через указанный разделитель
	AddFor(prefix, key, value, delim string)
	// Get возвращает значение опции конфига по ключу
	Get(key string) (value string, ok bool)
	// GetFor возвращает значение опции для префикса по ключу
	GetFor(prefix, key string) (value string, ok bool)
	// DumpEnvFor возвращает содержимое конфига для префикса в виде слайса ключ=значение
	DumpEnvFor(prefix string, filter ...string) []string
	// DumpEnv возвращает содержимое конфига в виде слайса ключ=значение
	DumpEnv(filter ...string) []string
}

type localConfigController struct {
	server *http.Server
	sock   net.Listener

	sync.RWMutex
	m map[string]interface{}
}

// LocalConfigController - контроллер конфигурации по умолчанию
func LocalConfigController(ctx context.Context, log slog.Logger) (Controller, error) {
	ctl := &localConfigController{
		m: make(map[string]interface{}),
	}

	sock, err := net.Listen("tcp", ":0") // nolint
	if err != nil {
		return nil, errors.Wrap(err, "listen controller API server")
	}

	ctl.sock = sock
	ctl.server = &http.Server{Handler: ctl.router()}

	_, port, _ := net.SplitHostPort(sock.Addr().String())
	ctl.m["CONFIG_CONTROLLER_PORT"] = port

	go func() {
		if err = ctl.server.Serve(ctl.sock); err != nil {
			log.Error(ctx, "start controller API server")
		}
	}()

	return ctl, nil
}

// Endpoint - возвращает адрес сервера контроллера
func (ctl *localConfigController) Endpoint() string {
	if ctl == nil {
		return ""
	}

	return ctl.sock.Addr().String()
}

// Set устанавливает опцию конфига
func (ctl *localConfigController) Set(key, value string) {
	if ctl == nil {
		return
	}

	ctl.Lock()
	defer ctl.Unlock()

	ctl.m[key] = value
}

// SetFor устанавливает опцию конфига для префикса
func (ctl *localConfigController) SetFor(prefix, key, value string) {
	if ctl == nil {
		return
	}

	pm := ctl.getPrefixedMap(prefix)

	ctl.Lock()
	defer ctl.Unlock()

	pm[key] = value
	ctl.m[prefix] = pm
}

// Add - устанавливает опцию конфига, либо добавляет значение
// в существующую через указанный разделитель
func (ctl *localConfigController) Add(key, value, delim string) {
	if ctl == nil {
		return
	}

	ctl.Lock()
	defer ctl.Unlock()

	setVal := ""

	if v, ok := ctl.m[key].(string); ok {
		setVal = v + delim
	}

	ctl.m[key] = setVal + value
}

// AddFor устанавливает опцию конфига для префикса, либо добавляет значение
// в существующую через указанный разделитель
func (ctl *localConfigController) AddFor(prefix, key, value, delim string) {
	if ctl == nil {
		return
	}

	pm := ctl.getPrefixedMap(prefix)

	ctl.Lock()
	defer ctl.Unlock()

	setVal := ""

	if v, exist := pm[key].(string); exist {
		setVal = v + delim
	}

	pm[key] = setVal + value
	ctl.m[prefix] = pm
}

// Get возвращает значение опции конфига по ключу
func (ctl *localConfigController) Get(key string) (value string, ok bool) {
	if ctl == nil {
		return "", false
	}

	ctl.RLock()
	defer ctl.RUnlock()

	if value, ok = ctl.m[key].(string); ok {
		return value, true
	}

	return "", false
}

// GetFor возвращает значение опции для префикса по ключу
func (ctl *localConfigController) GetFor(prefix, key string) (value string, ok bool) {
	if ctl == nil {
		return "", false
	}

	pm := ctl.getPrefixedMap(prefix)

	ctl.RLock()
	defer ctl.RUnlock()

	if value, ok = pm[key].(string); ok {
		return value, true
	}

	return "", false
}

// DumpEnv возвращает содержимое конфига в виде слайса ключ=значение
func (ctl *localConfigController) DumpEnv(filter ...string) []string {
	if ctl == nil {
		return nil
	}

	ctl.RLock()
	defer ctl.RUnlock()

	dump := make([]string, 0, len(ctl.m))

	if len(filter) == 0 {
		for k, v := range ctl.m {
			if val, ok := v.(string); ok {
				dump = append(dump, fmt.Sprintf("%s=%s", k, val))
			}
		}
	} else {
		for _, k := range filter {
			if v, ok := ctl.m[k].(string); ok {
				dump = append(dump, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}

	return dump
}

// DumpEnvFor возвращает содержимое конфига для префикса в виде слайса ключ=значение
func (ctl *localConfigController) DumpEnvFor(prefix string, filter ...string) []string {
	if ctl == nil {
		return nil
	}

	pm := ctl.getPrefixedMap(prefix)

	if len(pm) == 0 {
		return nil
	}

	dump := make([]string, 0, len(pm))

	if len(filter) == 0 {
		for k, v := range pm {
			if val, ok := v.(string); ok {
				dump = append(dump, fmt.Sprintf("%s=%s", k, val))
			}
		}
	} else {
		for _, k := range filter {
			if v, ok := pm[k].(string); ok {
				dump = append(dump, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}

	return dump
}

// Close завершает работу интерфейса для удаленного подключения к контроллеру
func (ctl *localConfigController) Close() error {
	if ctl == nil {
		return nil
	}

	if err := ctl.server.Shutdown(context.Background()); err != nil {
		return errors.Wrap(err, "API server shutdown")
	}

	return nil
}

func (ctl *localConfigController) getPrefixedMap(prefix string) map[string]interface{} {
	ctl.RLock()
	defer ctl.RUnlock()

	if pm, ok := ctl.m[prefix].(map[string]interface{}); ok {
		return pm
	}

	return make(map[string]interface{})
}

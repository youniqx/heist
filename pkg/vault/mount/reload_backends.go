package mount

import (
	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type reloadScope string

const (
	reloadScopeGlobal reloadScope = "global"
)

type reloadRequest struct {
	Plugin Plugin      `json:"plugin,omitempty"`
	Mounts []string    `json:"mounts,omitempty"`
	Scope  reloadScope `json:"scope,omitempty"`
}

func (a *mountAPI) ReloadPluginBackends(plugin Plugin) error {
	log := a.Core.Log().WithValues("method", "ReloadPluginBackends", "plugin", plugin)

	request := &reloadRequest{
		Plugin: plugin,
		Mounts: nil,
		Scope:  reloadScopeGlobal,
	}

	if err := a.Core.MakeRequest(core.MethodPut, "/v1/sys/plugins/reload/backend", httpclient.JSON(request), nil); err != nil {
		log.Info("failed to reload plugin backends", "error", err)
		return core.ErrAPIError.WithDetails("failed to reload plugin backends").WithCause(err)
	}

	return nil
}

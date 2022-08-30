package mount

import (
	"strings"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type listMountsResponse struct {
	Data map[string]*mountInfo `json:"data"`
}

type mountInfo struct {
	UUID        string            `json:"uuid,omitempty"`
	Type        Type              `json:"type,omitempty"`
	Description string            `json:"description,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
	Config      *TuneConfig       `json:"config,omitempty"`
}

func (a *mountAPI) ListMounts() ([]*Mount, error) {
	log := a.Core.Log().WithValues("method", "ListMounts")

	response := &listMountsResponse{}
	if err := a.Core.MakeRequest(core.MethodGet, "/v1/sys/mounts", nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to list mounted engines", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to list mounts").WithCause(err)
	}

	engines := make([]*Mount, 0, len(response.Data))

	for path, mount := range response.Data {
		if mount.Type == TypeKVV1 && mount.Options["version"] == "2" {
			mount.Type = TypeKVV2
		}

		var options map[string]string
		if len(mount.Options) == 0 {
			options = nil
		} else {
			options = mount.Options
		}

		engines = append(engines, &Mount{
			Path:    strings.TrimSuffix(path, "/"),
			Type:    mount.Type,
			Options: options,
			Config:  mount.Config,
		})
	}

	return engines, nil
}

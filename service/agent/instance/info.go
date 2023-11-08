package instance

import (
	"context"
	"encoding/json"
	"net/http"
	"tinyvm/internal"
	"tinyvm/service/agent/params"

	lxd "github.com/canonical/lxd/client"
)

func GetInstanceInfo(ctx context.Context) http.HandlerFunc {
	lxd_server_ctx := ctx.Value(internal.CTX_LXD_SERVER)
	lxd_server := lxd_server_ctx.(lxd.InstanceServer)
	return func(w http.ResponseWriter, r *http.Request) {
		param := params.DeleteInstanceParam{}
		if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 400001,
				Msg:  "参数错误",
				Data: err.Error(),
			})
			return
		}
		instance, _, err := lxd_server.GetInstanceState(param.Cid)
		if err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 400002,
				Msg:  "获取实例信息失败",
				Data: err.Error(),
			})
			return
		}
		storage, err := lxd_server.GetStoragePoolResources("test111")
		if err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 400003,
				Msg:  "获取存储信息失败",
				Data: err.Error(),
			})
			return
		}
		resp := params.InstanceInfoResp{}
		resp.Memory = instance.Memory
		for i := range instance.Network {
			resp.Network.BytesReceived += instance.Network[i].Counters.BytesReceived
			resp.Network.BytesSent += instance.Network[i].Counters.BytesSent
			resp.Network.PacketsReceived += instance.Network[i].Counters.PacketsReceived
			resp.Network.PacketsSent += instance.Network[i].Counters.PacketsSent
			resp.Network.ErrorsReceived += instance.Network[i].Counters.ErrorsReceived
			resp.Network.ErrorsSent += instance.Network[i].Counters.ErrorsSent
			resp.Network.PacketsDroppedInbound += instance.Network[i].Counters.PacketsDroppedInbound
			resp.Network.PacketsDroppedOutbound += instance.Network[i].Counters.PacketsDroppedOutbound
		}
		resp.Disk = storage.Space
		resp.State = instance.Status
		_ = json.NewEncoder(w).Encode(params.BaseResp[params.InstanceInfoResp]{
			Code: 0,
			Msg:  "获取成功",
			Data: resp,
		})
	}
}

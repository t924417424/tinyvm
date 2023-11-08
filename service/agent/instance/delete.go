package instance

import (
	"context"
	"encoding/json"
	"net/http"
	"tinyvm/internal"
	"tinyvm/service/agent/params"

	lxd "github.com/canonical/lxd/client"
)

func DeleteInstance(ctx context.Context) http.HandlerFunc {
	lxd_server_ctx := ctx.Value(internal.CTX_LXD_SERVER)
	lxd_server := lxd_server_ctx.(lxd.InstanceServer)
	return func(w http.ResponseWriter, r *http.Request) {
		param := params.DeleteInstanceParam{}
		if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 300001,
				Msg:  "参数错误",
				Data: err.Error(),
			})
			return
		}
		// 删除实例
		op, err := lxd_server.DeleteInstance(param.Cid)
		if err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 300002,
				Msg:  "操作失败",
				Data: err.Error(),
			})
			return
		}
		if err := op.Wait(); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 300003,
				Msg:  "删除实例失败",
				Data: err.Error(),
			})
			return
		}
		if err := lxd_server.DeleteNetwork(param.Cid); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 300004,
				Msg:  "删除实例网络失败",
				Data: err.Error(),
			})
			return
		}
		if err := lxd_server.DeleteStoragePool(param.Cid); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 300005,
				Msg:  "删除实例存储卷失败",
				Data: err.Error(),
			})
			return
		}
		if err := lxd_server.DeleteNetwork(param.Cid); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[struct{}]{
				Code: 0,
				Msg:  "删除实例成功",
			})
			return
		}
	}
}

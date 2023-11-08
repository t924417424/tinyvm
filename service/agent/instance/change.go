package instance

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"tinyvm/internal"
	"tinyvm/service/agent/params"

	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
)

func ChangePasswd(ctx context.Context) http.HandlerFunc {
	lxd_server_ctx := ctx.Value(internal.CTX_LXD_SERVER)
	lxd_server := lxd_server_ctx.(lxd.InstanceServer)
	return func(w http.ResponseWriter, r *http.Request) {
		param := params.ChangePasswdParam{}
		if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 200001,
				Msg:  "参数错误",
				Data: err.Error(),
			})
			return
		}
		inc, _, err := lxd_server.GetInstance(param.Cid)
		if err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 200002,
				Msg:  "获取实例信息失败",
				Data: err.Error(),
			})
			return
		}
		if inc.StatusCode != api.Running {
			_ = json.NewEncoder(w).Encode(params.BaseResp[struct{}]{
				Code: 200003,
				Msg:  "实例非运行状态",
			})
			return
		}
		stdInOut := internal.NewExecIO(func(in []byte) []byte {
			if strings.Contains(string(in), "password") {
				return []byte(param.NewPassword)
			}
			return nil
		})
		done := make(chan bool)
		op, err := lxd_server.ExecInstance(param.Cid, api.InstanceExecPost{
			Command: []string{"passwd"},
			// Interactive: true,
			WaitForWS: true,
		}, &lxd.InstanceExecArgs{
			Stdout:   stdInOut,
			Stdin:    stdInOut,
			DataDone: done,
		})
		if err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 200004,
				Msg:  "实例操作失败失败",
				Data: err.Error(),
			})
			return
		}
		if err := op.Wait(); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 200005,
				Msg:  "密码修改操作失败失败",
				Data: err.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(params.BaseResp[struct{}]{
			Code: 0,
			Msg:  "修改成功",
		})
	}
}

func ChangeState(ctx context.Context) http.HandlerFunc {
	lxd_server_ctx := ctx.Value(internal.CTX_LXD_SERVER)
	lxd_server := lxd_server_ctx.(lxd.InstanceServer)
	return func(w http.ResponseWriter, r *http.Request) {
		param := params.ChangeInstanceState{}
		if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 210001,
				Msg:  "参数错误",
				Data: err.Error(),
			})
			return
		}
		inc, _, err := lxd_server.GetInstance(param.Cid)
		if err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 210002,
				Msg:  "获取实例信息失败",
				Data: err.Error(),
			})
			return
		}
		_ = inc
		op, err := lxd_server.UpdateInstanceState(param.Cid, api.InstanceStatePut{
			Action:  param.State,
			Timeout: 10,
			Force:   param.Force,
		}, "")
		if err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 210003,
				Msg:  "操作实例失败",
				Data: err.Error(),
			})
			return
		}
		if err := op.Wait(); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 210004,
				Msg:  "更改实例状态失败",
				Data: err.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(params.BaseResp[struct{}]{
			Code: 0,
			Msg:  "操作成功",
		})
	}
}

// func ChangeNetwork(ctx context.Context) http.HandlerFunc {
// 	lxd_server_ctx := ctx.Value(internal.CTX_LXD_SERVER)
// 	lxd_server := lxd_server_ctx.(lxd.InstanceServer)
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		param := params.ChangePasswdParam{}
// 		if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
// 			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
// 				Code: 200001,
// 				Msg:  "参数错误",
// 				Data: err.Error(),
// 			})
// 			return
// 		}
// 		inc, _, err := lxd_server.GetInstance(param.Cid)
// 		if err != nil {
// 			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
// 				Code: 200002,
// 				Msg:  "获取实例信息失败",
// 				Data: err.Error(),
// 			})
// 			return
// 		}
// 		_ = inc
// 		// lxd_server.UpdateInstance("",api.InstancePut{
// 		// 	Config:       map[string]string{},
// 		// })
// 	}
// }

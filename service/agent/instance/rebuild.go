package instance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"tinyvm/internal"
	"tinyvm/service/agent/params"

	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
)

func RebuildInstanceStatus(ctx context.Context) http.HandlerFunc {
	lxd_server_ctx := ctx.Value(internal.CTX_LXD_SERVER)
	lxd_server := lxd_server_ctx.(lxd.InstanceServer)
	return func(w http.ResponseWriter, r *http.Request) {
		param := params.CreateInstanceParams{}
		if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 600001,
				Msg:  "参数错误",
				Data: err.Error(),
			})
			return
		}
		if err := param.Check(); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 600002,
				Msg:  "参数检查不通过",
				Data: err.Error(),
			})
			return
		}
		// ns, _, err := lxd_server.GetNetwork(param.Cid)
		// if err != nil {
		// 	_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
		// 		Code: 600003,
		// 		Msg:  "原始IP操作失败",
		// 		Data: err.Error(),
		// 	})
		// 	return
		// }
		// // rawIp, ok := ns.Config["ipv4.address"]
		// if !ok {
		// 	_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
		// 		Code: 600004,
		// 		Msg:  "原始IP不存在",
		// 		Data: err.Error(),
		// 	})
		// 	return
		// }
		op, err := lxd_server.UpdateInstanceState(param.Cid, api.InstanceStatePut{
			Action: "stop",
		}, "")
		if err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 600005,
				Msg:  "关机操作失败",
				Data: err.Error(),
			})
			return
		}
		if err := op.Wait(); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 600006,
				Msg:  "关机失败",
				Data: err.Error(),
			})
			return
		}
		op, err = lxd_server.DeleteInstance(param.Cid)
		if err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 600007,
				Msg:  "删除原实例操作失败",
				Data: err.Error(),
			})
			return
		}
		if err := op.Wait(); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 600008,
				Msg:  "删除原实例失败",
				Data: err.Error(),
			})
			return
		}
		devices := map[string]map[string]string{
			"root": {
				"type": "disk",
				"pool": param.Cid,
				"path": "/",
			},
			"eth0": {
				"name":    "eth0",
				"nictype": "bridged",
				"parent":  param.Cid,
				"type":    "nic",
			},
			// 需要开启动ndp相关
			// sysctl net.ipv6.conf.all.proxy_ndp=1
			// sysctl net.ipv6.conf.eth0.proxy_ndp=1
			// sysctl net.ipv6.conf.all.forwarding = 1
		}
		for i := range param.Ipv6 {
			devices[fmt.Sprintf("eth%d", i+1)] = map[string]string{
				"name":         fmt.Sprintf("eth%d", i+1),
				"nictype":      "routed",
				"parent":       "eth0", // 注意，这个是宿主机的ipv6网卡名
				"type":         "nic",
				"ipv6.address": param.Ipv6[i],
			}
		}
		// 创建实例
		op, err = lxd_server.CreateInstance(api.InstancesPost{
			InstancePut: api.InstancePut{
				Config: map[string]string{
					"limits.cpu":           fmt.Sprintf("%d", param.Limits.CPU.Core),
					"limits.memory":        fmt.Sprintf("%dMB", param.Limits.RAM.Memory),
					"limits.cpu.allowance": fmt.Sprintf("%dms/100ms", param.Limits.CPU.Allowance),
				},
				Devices: devices,
			},
			Name: param.Cid,
			// 如果源有更新则会拉取，建议拉取和创建分开操作
			Source: api.InstanceSource{
				Type:     param.Os.Type,        // image
				Server:   param.Os.ImageServer, // https://mirrors.tuna.tsinghua.edu.cn/lxc-images/
				Alias:    param.Os.Name,        // alpine/3.16
				Protocol: "simplestreams",
			},
			Type: api.InstanceTypeContainer,
		})
		if err != nil {
			// _ = lxd_server.DeleteStoragePool(param.Cid)
			// _ = lxd_server.DeleteNetwork(param.Cid)
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100005,
				Msg:  "实例创建失败",
				Data: err.Error(),
			})
			return
		}
		if err := op.Wait(); err != nil {
			// _ = lxd_server.DeleteStoragePool(param.Cid)
			// _ = lxd_server.DeleteNetwork(param.Cid)
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100006,
				Msg:  "实例初始化失败",
				Data: err.Error(),
			})
			return
		}
		op, err = lxd_server.UpdateInstanceState(param.Cid, api.InstanceStatePut{
			Action:  "start",
			Timeout: 10,
		}, "")
		if err != nil {
			// _ = lxd_server.DeleteStoragePool(param.Cid)
			// _ = lxd_server.DeleteNetwork(param.Cid)
			// op, _ = lxd_server.DeleteInstance(param.Cid)
			// _ = op.Wait()
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100007,
				Msg:  "实例开机操作失败",
				Data: err.Error(),
			})
			return
		}
		if err := op.Wait(); err != nil {
			// _ = lxd_server.DeleteStoragePool(param.Cid)
			// _ = lxd_server.DeleteNetwork(param.Cid)
			// op, _ = lxd_server.DeleteInstance(param.Cid)
			// _ = op.Wait()
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100008,
				Msg:  "实例开机失败",
				Data: err.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(params.BaseResp[struct{}]{
			Code: 0,
			Msg:  "实例创建成功",
		})
	}
}

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

func CreateInstance(ctx context.Context) http.HandlerFunc {
	lxd_server_ctx := ctx.Value(internal.CTX_LXD_SERVER)
	lxd_server := lxd_server_ctx.(lxd.InstanceServer)
	v4cidr := ctx.Value(internal.CTX_IPV4CIDI).(string)
	usedv4 := ctx.Value(internal.CTX_IPV4CIDI).([]string)
	return func(w http.ResponseWriter, r *http.Request) {
		param := params.CreateInstanceParams{}
		if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100001,
				Msg:  "参数错误",
				Data: err.Error(),
			})
			return
		}
		if err := param.Check(); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100002,
				Msg:  "参数检查不通过",
				Data: err.Error(),
			})
			return
		}
		createSuccess := false
		defer func() {
			// 如果创建虚拟机失败
			// free res
			if !createSuccess {
				_ = lxd_server.DeleteStoragePool(param.Cid)
				_ = lxd_server.DeleteNetwork(param.Cid)
				op, err := lxd_server.DeleteInstance(param.Cid)
				if err != nil {
					return
				}
				_ = op.Wait()
			}
		}()
		if err := lxd_server.CreateStoragePool(api.StoragePoolsPost{
			StoragePoolPut: api.StoragePoolPut{
				Config: map[string]string{
					"size": fmt.Sprintf("%dMiB", param.Limits.Disk.Size),
				},
			},
			Name:   param.Cid,
			Driver: "btrfs",
		}); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100003,
				Msg:  "存储卷创建失败",
				Data: err.Error(),
			})
			return
		}
		freeIp, err := internal.GenerateIP(v4cidr, usedv4)
		if err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100007,
				Msg:  "内网空闲IP用尽",
				Data: err.Error(),
			})
			return
		}
		if err := lxd_server.CreateNetwork(api.NetworksPost{
			Name: param.Cid,
			NetworkPut: api.NetworkPut{
				Config: map[string]string{
					"ipv4.address": freeIp,
					"ipv4.nat":     "true",
				},
				Description: "",
			},
			// 	Type: "bridge",
		}); err != nil {
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100004,
				Msg:  "实例网络创建失败",
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
		op, err := lxd_server.CreateInstance(api.InstancesPost{
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
			_ = lxd_server.DeleteStoragePool(param.Cid)
			_ = lxd_server.DeleteNetwork(param.Cid)
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100005,
				Msg:  "实例创建失败",
				Data: err.Error(),
			})
			return
		}
		if err := op.Wait(); err != nil {
			_ = lxd_server.DeleteStoragePool(param.Cid)
			_ = lxd_server.DeleteNetwork(param.Cid)
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
			_ = lxd_server.DeleteStoragePool(param.Cid)
			_ = lxd_server.DeleteNetwork(param.Cid)
			op, _ = lxd_server.DeleteInstance(param.Cid)
			_ = op.Wait()
			_ = json.NewEncoder(w).Encode(params.BaseResp[string]{
				Code: 100007,
				Msg:  "实例开机失败",
				Data: err.Error(),
			})
			return
		}
		if err := op.Wait(); err != nil {
			_ = lxd_server.DeleteStoragePool(param.Cid)
			_ = lxd_server.DeleteNetwork(param.Cid)
			op, _ = lxd_server.DeleteInstance(param.Cid)
			_ = op.Wait()
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

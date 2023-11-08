package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"tinyvm/internal"
	"tinyvm/service/agent/instance"
	"tinyvm/service/agent/middleware"

	lxd "github.com/canonical/lxd/client"
)

const (
	MIN_LXD_VERSION = "3.2.0"
)

var (
	addr     = ":8018"
	secret   = ""
	netinte  = "eth0"
	ipv4cidr = "10.1.0.1:24"
)

func main() {
	// logger.InitLogger("", "", false, true, nil)
	cli, err := lxd.ConnectLXDUnix("", &lxd.ConnectionArgs{})
	if err != nil {
		cli, err = lxd.ConnectLXDUnix("/var/snap/lxd/common/lxd/unix.socket", &lxd.ConnectionArgs{})
		if err != nil {
			panic(err)
		}
	}
	defer cli.Disconnect()
	ser, _, _ := cli.GetServer()
	println(ser.Environment.ServerVersion)
	if ser.Environment.ServerVersion < MIN_LXD_VERSION {
		fmt.Printf("最低支持LXD版本：%s 。\n", MIN_LXD_VERSION)
		os.Exit(1)
	}

	flag.StringVar(&addr, "listen", ":8018", "agent端监听地址")
	flag.StringVar(&secret, "secret", "", "agent通信密钥")
	flag.StringVar(&netinte, "netinte", "eth0", "使用的网卡")
	flag.StringVar(&ipv4cidr, "ipv4cidr", "10.1.0.1:24", "内网IPV4分配范围")
	flag.Parse()

	var usedIPV4 = make([]string, 0)
	ns, err := cli.GetNetworks()
	if err != nil {
		panic(err.Error())
	}
	for i := range ns {
		ni, _, err := cli.GetNetwork(ns[i].Name)
		if err != nil {
			panic(err.Error())
		}
		if ip, ok := ni.Config["ipv4.address"]; ok {
			usedIPV4 = append(usedIPV4, ip)
		}
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, internal.CTX_IPV4CIDI, ipv4cidr)
	ctx = context.WithValue(ctx, internal.CTX_SECRET, secret)
	ctx = context.WithValue(ctx, internal.CTX_USED_IPV4, usedIPV4)
	ctx = context.WithValue(ctx, internal.CTX_LXD_SERVER, cli)
	http.HandleFunc("/instence/create", middleware.Protect(ctx, instance.CreateInstance))
	http.HandleFunc("/instence/delete", middleware.Protect(ctx, instance.DeleteInstance))
	http.HandleFunc("/instence/change_passwd", middleware.Protect(ctx, instance.ChangePasswd))
	http.HandleFunc("/instence/change_state", middleware.Protect(ctx, instance.ChangeState))
	if err := http.ListenAndServe(":8018", nil); err != nil {
		fmt.Printf("服务启动失败：%s\n", err.Error())
		os.Exit(1)
	}

	// 创建存储池
	// err = cli.CreateStoragePool(api.StoragePoolsPost{
	// 	Name: "test-api-1",
	// 	StoragePoolPut: api.StoragePoolPut{
	// 		Config: map[string]string{
	// 			"size": "128MB",
	// 		},
	// 	},
	// 	Driver: "btrfs",
	// })
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }
	// 创建网络
	// err = cli.CreateNetwork(api.NetworksPost{
	// 	Name: "test-api-1",
	// 	NetworkPut: api.NetworkPut{
	// 		Config: map[string]string{
	// 			"ipv4.address": "auto",
	// 			"ipv4.nat":     "true",
	// 		},
	// 		Description: "",
	// 	},
	// 	Type: "bridge",
	// })
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }
	// if err != nil {
	// 	println(err.Error())
	// }
	// Connect to the remote SimpleStreams server
	// d, err := lxd.ConnectSimpleStreams("https://mirrors.tuna.tsinghua.edu.cn/lxc-images/", nil)
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }
	// defer d.Disconnect()
	// img, _, err := d.GetImageAlias("alpine/3.18")
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }
	// println(img.Name)
	// Get the image information
	// image, _, err := d.GetImage(img.Target)
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }
	// 创建实例
	// op, err := cli.CreateInstance(api.InstancesPost{
	// 	InstancePut: api.InstancePut{
	// 		Devices: map[string]map[string]string{
	// 			"root": {
	// 				"type": "disk",
	// 				"pool": "test-api-1",
	// 				"path": "/",
	// 			},
	// 			"eth0": {
	// 				"name":    "eth0",
	// 				"nictype": "bridged",
	// 				"parent":  "test-api-1",
	// 				"type":    "nic",
	// 			},
	// 			// 需要开启动ndp相关
	// 			// sysctl net.ipv6.conf.all.proxy_ndp=1
	// 			// sysctl net.ipv6.conf.eth0.proxy_ndp=1
	// 			// sysctl net.ipv6.conf.all.forwarding = 1
	// 			"eth1": {
	// 				"name":         "eth1",
	// 				"nictype":      "routed",
	// 				"parent":       "eth0", // 注意，这个是宿主机的ipv6网卡名
	// 				"type":         "nic",
	// 				"ipv6.address": "2402:d0c0:18:62fd:216:3eff:fe0e:4c9b",
	// 			},
	// 		},
	// 	},
	// 	Name: "test-api-1",
	// 	// 如果源有更新则会拉取，建议拉取和创建分开操作
	// 	Source: api.InstanceSource{
	// 		Type:     "image",
	// 		Server:   "https://mirrors.tuna.tsinghua.edu.cn/lxc-images/",
	// 		Alias:    "alpine/3.16",
	// 		Protocol: "simplestreams",
	// 	},
	// 	Type: api.InstanceTypeContainer,
	// })
	// 先获取原来的才能更新
	// op, err := cli.UpdateInstance("test-api5", api.InstancePut{
	// 	Devices: map[string]map[string]string{
	// 		"root": {
	// 			"type": "disk",
	// 			"pool": "test-api",
	// 			"path": "/",
	// 		},
	// 		"eth1": {
	// 			"name":         "eth1",
	// 			"nictype":      "routed",
	// 			"parent":       "eth0", // 注意，这个是宿主机的ipv6网卡名
	// 			"type":         "nic",
	// 			"ipv6.address": "fe80::215:5dff:fe99:c02f",
	// 		},
	// 	},
	// }, "")
	// op, err := cli.UpdateInstanceState("test-api4", api.InstanceStatePut{
	// 	Action:  "start",
	// 	Timeout: 10,
	// }, "")
	// if err != nil {
	// 	panic(err)
	// }
	// println(op.Get().ID)
	// err = op.Wait()
	// if err != nil {
	// 	panic(err)
	// }
	// stdout := new(wsBufferWriter)
	// newpwd := []byte("123456")
	// newpwd = append(newpwd, 10)
	// stdInOut := NewExecIO(func(in []byte) []byte {
	// 	println(string(in))
	// 	if strings.Contains(string(in), "password") {
	// 		return newpwd
	// 	}
	// 	return nil
	// })
	// done := make(chan bool)
	// op, err := cli.ExecInstance("test-api6", api.InstanceExecPost{
	// 	Command: []string{"passwd"},
	// 	// Interactive: true,
	// 	WaitForWS: true,
	// }, &lxd.InstanceExecArgs{
	// 	Stdout:   stdInOut,
	// 	Stdin:    stdInOut,
	// 	DataDone: done,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// <-done
	// i, _, err := cli.GetInstance("test-api6")
	// if err != nil {
	// 	panic(err)
	// }
	// i.Devices["eth0"]["limits.max"] = "1Mbit"
	// op, err := cli.UpdateInstance("test-api6", api.InstancePut{
	// 	Config:  i.Config,
	// 	Devices: i.Devices,
	// }, "")
	// if err != nil {
	// 	panic(err)
	// }
	// err = op.Wait()
	// if err != nil {
	// 	panic(err)
	// }

	// param := params.CreateInstanceParams{
	// 	ActionId: "test",
	// 	Cid:      "test111",
	// 	Limits: params.ResLimit{
	// 		Disk: params.DiskLimit{
	// 			Size: 128,
	// 		},
	// 		BandWidth: params.NetworkLimit{
	// 			Rate: 100,
	// 		},
	// 		RAM: params.RAMLimit{
	// 			Memory: 64,
	// 		},
	// 		CPU: params.CPULimit{
	// 			Core:      1,
	// 			Allowance: 100,
	// 		},
	// 	},
	// 	Os: params.OsInfo{
	// 		Type:        "image",
	// 		ImageServer: "https://mirrors.tuna.tsinghua.edu.cn/lxc-images/",
	// 		Name:        "alpine/3.15",
	// 		Version:     "",
	// 		Arch:        "",
	// 	},
	// 	Ipv6: []string{
	// 		"fe80::215:5dff:fe99:c02f",
	// 	},
	// 	Ipv4: []string{
	// 		"10.0.1.1/24",
	// 	},
	// 	Privileged: false,
	// }
	// op, err := cli.UpdateInstanceState(param.Cid, api.InstanceStatePut{
	// 	Action: "stop",
	// }, "")
	// if err != nil {
	// 	panic(err)
	// 	return
	// }
	// if err := op.Wait(); err != nil {
	// 	panic(err)
	// 	return
	// }
	// op, err = cli.DeleteInstance(param.Cid)
	// if err != nil {
	// 	panic(err)
	// 	return
	// }
	// if err := op.Wait(); err != nil {
	// 	panic(err)
	// 	return
	// }
	// devices := map[string]map[string]string{
	// 	"root": {
	// 		"type": "disk",
	// 		"pool": param.Cid,
	// 		"path": "/",
	// 	},
	// 	"eth0": {
	// 		"name":    "eth0",
	// 		"nictype": "bridged",
	// 		"parent":  param.Cid,
	// 		"type":    "nic",
	// 	},
	// 	// 需要开启动ndp相关
	// 	// sysctl net.ipv6.conf.all.proxy_ndp=1
	// 	// sysctl net.ipv6.conf.eth0.proxy_ndp=1
	// 	// sysctl net.ipv6.conf.all.forwarding = 1
	// }
	// for i := range param.Ipv6 {
	// 	devices[fmt.Sprintf("eth%d", i+1)] = map[string]string{
	// 		"name":         fmt.Sprintf("eth%d", i+1),
	// 		"nictype":      "routed",
	// 		"parent":       "eth0", // 注意，这个是宿主机的ipv6网卡名
	// 		"type":         "nic",
	// 		"ipv6.address": param.Ipv6[i],
	// 	}
	// }
	// // 创建实例
	// op, err = cli.CreateInstance(api.InstancesPost{
	// 	InstancePut: api.InstancePut{
	// 		Config: map[string]string{
	// 			"limits.cpu":           fmt.Sprintf("%d", param.Limits.CPU.Core),
	// 			"limits.memory":        fmt.Sprintf("%dMB", param.Limits.RAM.Memory),
	// 			"limits.cpu.allowance": fmt.Sprintf("%dms/100ms", param.Limits.CPU.Allowance),
	// 		},
	// 		Devices: devices,
	// 	},
	// 	Name: param.Cid,
	// 	// 如果源有更新则会拉取，建议拉取和创建分开操作
	// 	Source: api.InstanceSource{
	// 		Type:     param.Os.Type,        // image
	// 		Server:   param.Os.ImageServer, // https://mirrors.tuna.tsinghua.edu.cn/lxc-images/
	// 		Alias:    param.Os.Name,        // alpine/3.16
	// 		Protocol: "simplestreams",
	// 	},
	// 	Type: api.InstanceTypeContainer,
	// })
	// if err != nil {
	// 	panic(err)
	// 	return
	// }
	// if err := op.Wait(); err != nil {
	// 	panic(err)
	// 	return
	// }
	// op, err = cli.UpdateInstanceState(param.Cid, api.InstanceStatePut{
	// 	Action:  "start",
	// 	Timeout: 10,
	// }, "")
	// if err != nil {
	// 	panic(err)
	// 	return
	// }
	// if err := op.Wait(); err != nil {
	// 	panic(err)
	// 	return
	// }
	// createSuccess := false
	// defer func() {
	// 	// free res
	// 	if !createSuccess {
	// 		_ = cli.DeleteStoragePool(param.Cid)
	// 		_ = cli.DeleteNetwork(param.Cid)
	// 		op, err := cli.DeleteInstance(param.Cid)
	// 		if err != nil {
	// 			return
	// 		}
	// 		_ = op.Wait()
	// 	}
	// }()
	// if err := cli.CreateStoragePool(api.StoragePoolsPost{
	// 	StoragePoolPut: api.StoragePoolPut{
	// 		Config: map[string]string{
	// 			"size": fmt.Sprintf("%dMiB", param.Limits.Disk.Size),
	// 		},
	// 	},
	// 	Name:   param.Cid,
	// 	Driver: "btrfs",
	// }); err != nil {
	// 	println(err.Error())
	// 	return
	// }
	// if err := cli.CreateNetwork(api.NetworksPost{
	// 	Name: param.Cid,
	// 	NetworkPut: api.NetworkPut{
	// 		Config: map[string]string{
	// 			"ipv4.address": param.Ipv4[0],
	// 			"ipv4.nat":     "true",
	// 		},
	// 		Description: "",
	// 	},
	// 	// 	Type: "bridge",
	// }); err != nil {
	// 	println(err.Error())
	// 	return
	// }
	// devices := map[string]map[string]string{
	// 	"root": {
	// 		"type": "disk",
	// 		"pool": param.Cid,
	// 		"path": "/",
	// 	},
	// 	"eth0": {
	// 		"name":    "eth0",
	// 		"nictype": "bridged",
	// 		"parent":  param.Cid,
	// 		"type":    "nic",
	// 	},
	// 	// 需要开启动ndp相关
	// 	// sysctl net.ipv6.conf.all.proxy_ndp=1
	// 	// sysctl net.ipv6.conf.eth0.proxy_ndp=1
	// 	// sysctl net.ipv6.conf.all.forwarding = 1
	// }
	// for i := range param.Ipv6 {
	// 	devices[fmt.Sprintf("eth%d", i+1)] = map[string]string{
	// 		"name":         fmt.Sprintf("eth%d", i+1),
	// 		"nictype":      "routed",
	// 		"parent":       "eth0", // 注意，这个是宿主机的ipv6网卡名
	// 		"type":         "nic",
	// 		"ipv6.address": param.Ipv6[i],
	// 	}
	// }
	// // 创建实例
	// op, err := cli.CreateInstance(api.InstancesPost{
	// 	InstancePut: api.InstancePut{
	// 		Config: map[string]string{
	// 			"limits.cpu":           fmt.Sprintf("%d", param.Limits.CPU.Core),
	// 			"limits.memory":        fmt.Sprintf("%dMB", param.Limits.RAM.Memory),
	// 			"limits.cpu.allowance": fmt.Sprintf("%dms/100ms", param.Limits.CPU.Allowance),
	// 		},
	// 		Architecture: param.Os.Arch,
	// 		Devices:      devices,
	// 	},
	// 	Name: param.Cid,
	// 	// 如果源有更新则会拉取，建议拉取和创建分开操作
	// 	Source: api.InstanceSource{
	// 		Type:     param.Os.Type,        // image
	// 		Server:   param.Os.ImageServer, // https://mirrors.tuna.tsinghua.edu.cn/lxc-images/
	// 		Alias:    param.Os.Name,        // alpine/3.16
	// 		Protocol: "simplestreams",
	// 	},
	// 	Type: api.InstanceTypeContainer,
	// })
	// if err != nil {
	// 	_ = cli.DeleteStoragePool(param.Cid)
	// 	_ = cli.DeleteNetwork(param.Cid)
	// 	println(err.Error())
	// 	return
	// }
	// if err := op.Wait(); err != nil {
	// 	_ = cli.DeleteStoragePool(param.Cid)
	// 	_ = cli.DeleteNetwork(param.Cid)
	// 	println(err.Error())
	// 	return
	// }
	// op, err = cli.UpdateInstanceState(param.Cid, api.InstanceStatePut{
	// 	Action:  "start",
	// 	Timeout: 10,
	// }, "")
	// if err != nil {
	// 	_ = cli.DeleteStoragePool(param.Cid)
	// 	_ = cli.DeleteNetwork(param.Cid)
	// 	op, _ = cli.DeleteInstance(param.Cid)
	// 	_ = op.Wait()
	// 	println(err.Error())
	// 	return
	// }
	// if err := op.Wait(); err != nil {
	// 	_ = cli.DeleteStoragePool(param.Cid)
	// 	_ = cli.DeleteNetwork(param.Cid)
	// 	op, _ = cli.DeleteInstance(param.Cid)
	// 	_ = op.Wait()
	// 	println(err.Error())
	// 	return
	// }
	// createSuccess = true
	// println("ok")
	// instance, _, err := cli.GetInstanceState("test111")
	// if err != nil {
	// 	panic(err.Error())
	// }
	// fmt.Printf("%#+v\n", instance.Memory)
	// fmt.Printf("%#+v\n", instance.Network["eth0"].Counters)
	// storage, err := cli.GetStoragePoolResources("test111")
	// if err != nil {
	// 	panic(err.Error())
	// }
	// fmt.Printf("%#+v\n", storage)
	// ns, _, err := cli.GetNetwork("test111")
	// if err != nil {
	// 	panic(err.Error())
	// }
	// fmt.Println(ns.Config["ipv4.address"])
}

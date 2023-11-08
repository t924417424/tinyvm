package params

import "github.com/canonical/lxd/shared/api"

type ResLimit struct {
	Disk      DiskLimit    `json:"disk"`
	BandWidth NetworkLimit `json:"bandwidth"`
	RAM       RAMLimit     `json:"ram"`
	CPU       CPULimit     `json:"cpu"`
}

type DiskLimit struct {
	Size uint64 `json:"size"`
}

type NetworkLimit struct {
	Rate uint64 `json:"rate"`
}

type RAMLimit struct {
	Memory uint64 `json:"memory"`
}

type CPULimit struct {
	Core      uint64 `json:"core"`
	Allowance uint64 `json:"allowance"`
}

type OsInfo struct {
	Type        string `json:"type"`
	ImageServer string `json:"image_server"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Arch        string `json:"arch"`
}

type CreateInstanceParams struct {
	ActionId string   `json:"action_id"` // 操作ID
	Cid      string   `json:"cid"`
	Limits   ResLimit `json:"limits"` //资源限制
	Os       OsInfo   `json:"os"`
	// privileged
	Ipv6       []string `json:"ipv6"`
	Ipv4       []string `json:"ipv4"`
	Privileged bool     `json:"privileged"`
}

func (c *CreateInstanceParams) Check() error {
	return nil
}

type InstanceBase struct {
	ActionId string `json:"action_id"` // 操作ID
	Cid      string `json:"cid"`
}

type ChangePasswdParam struct {
	InstanceBase
	NewPassword string `json:"new_password"`
}

type ChangeNetworkParam struct {
	InstanceBase
	Rate uint64 `json:"rate"`
}

type DeleteInstanceParam struct {
	InstanceBase
}

type GetInstanceStatusParam struct {
	InstanceBase
}

type ChangeInstanceState struct {
	InstanceBase
	// start, stop, restart, freeze, unfreeze
	State string `json:"state"`
	Force bool   `json:"force"`
}

type InstanceInfoResp struct {
	State   string                           `json:"state"`
	Disk    api.ResourcesStoragePoolSpace    `json:"disk"`
	Memory  api.InstanceStateMemory          `json:"memory"`
	Network api.InstanceStateNetworkCounters `json:"network"`
}

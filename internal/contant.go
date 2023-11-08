package internal

type ctxKey string

const (
	CTX_LXD_SERVER ctxKey = "__ctx.lxd.server__"
	CTX_SECRET     ctxKey = "__ctx.secret.key__"
	CTX_IPV4CIDI   ctxKey = "__ctx.cidr.v4__"
	CTX_USED_IPV4  ctxKey = "__ctx.used.ipv4"
)

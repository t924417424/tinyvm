package agent

import (
	"fmt"
	"net/http"
)

type agent struct {
	host string
	port uint
	key  string
}

func NewAgent(host string, port uint, key string) agent {
	return agent{
		host: host,
		port: port,
		key:  key,
	}
}

func (a agent) Lanuch() {
	http.ListenAndServe(fmt.Sprintf("%s:%d", a.host, a.port), nil)
}

func (a agent) Connect() {

}

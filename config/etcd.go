package config

import (
	"github.com/coreos/go-etcd/etcd"
)

type EtcdConfiger struct {
	Machines []string
	client
}

func NewEtcdConfiger(machines []string) {

}

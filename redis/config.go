package redis

import (
	"fmt"
	"net/url"
)

type Config struct {
	Replication ReplicationConfig
}

type ReplicationConfig struct {
	Master *ReplicationMasterConfig
	Slave  *ReplicationSlaveConfig
}

func (r ReplicationConfig) Role() ReplicationRole {
	if r.Master != nil {
		return ReplicationRoleMaster
	}
	return ReplicationRoleSlave
}

type ReplicationRole int

const (
	ReplicationRoleMaster ReplicationRole = iota
	ReplicationRoleSlave
)

func (r ReplicationRole) String() string {
	switch r {
	case ReplicationRoleMaster:
		return "master"
	case ReplicationRoleSlave:
		return "slave"
	}
	panic(fmt.Sprintf("unknown redis.ReplicationRole: %d", r))
}

type ReplicationMasterConfig struct {
	ReplID     string
	ReplOffset uint64
}

type ReplicationSlaveConfig struct {
	MasterHost string
	MasterPort uint64
}

func (c ReplicationSlaveConfig) MasterAddress() (*url.URL, error) {
	return url.Parse(fmt.Sprintf("%s:%d", c.MasterHost, c.MasterPort))
}

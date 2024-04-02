package redis

import "fmt"

type Config struct {
	Replication ReplicationConfig
}

type ReplicationConfig struct {
	Master *ReplicationMasterConfig
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
	ReplOffset uint
}

package redis

import "fmt"

type Config struct {
	Replication *ReplicationConfig
}

type ReplicationConfig struct {
	Role ReplicationRole
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

package redis

type Config struct {
	Replication ReplicationConfig
}

type ReplicationConfig struct {
	Role ReplicationRole
}

type ReplicationRole int

const (
	ReplicationRoleMaster ReplicationRole = iota
	ReplicationRoleSlave
)

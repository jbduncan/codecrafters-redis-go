package redis_test

import "github.com/codecrafters-io/redis-starter-go/redis"

var (
	masterRedisConfig = &redis.Config{
		Replication: redis.ReplicationConfig{
			Master: &redis.ReplicationMasterConfig{
				ReplID:     "some-repl-id",
				ReplOffset: 0,
			},
		},
	}
	masterRedisConfigWithOtherReplID = &redis.Config{
		Replication: redis.ReplicationConfig{
			Master: &redis.ReplicationMasterConfig{
				ReplID:     "some-other-repl-id",
				ReplOffset: 0,
			},
		},
	}
	slaveRedisConfig = &redis.Config{
		Replication: redis.ReplicationConfig{},
	}
	zeroValueRedisConfig = &redis.Config{}
)

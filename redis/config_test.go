package redis_test

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis"
	. "github.com/onsi/gomega"
)

func TestReplicationRole_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    redis.ReplicationRole
		want string
	}{
		{
			name: "ReplicationRoleMaster",
			r:    redis.ReplicationRoleMaster,
			want: "master",
		},
		{
			name: "ReplicationRoleSlave",
			r:    redis.ReplicationRoleSlave,
			want: "slave",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("unknown replication role: -1", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(func() { _ = redis.ReplicationRole(-1).String() }).
			To(PanicWith("unknown redis.ReplicationRole: -1"))
	})

	t.Run("unknown replication role: 42", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(func() { _ = redis.ReplicationRole(42).String() }).
			To(PanicWith("unknown redis.ReplicationRole: 42"))
	})
}

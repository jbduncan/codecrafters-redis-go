package redis

import (
	"fmt"
	"reflect"
	"time"
)

type Command interface {
	Run() string
}

type EchoCommand string

func (e EchoCommand) Run() string {
	return bulkString(string(e))
}

func NewGetCommand(store *Store, clock Clock, key string) *GetCommand {
	return &GetCommand{
		store: store,
		clock: clock,
		key:   key,
	}
}

type GetCommand struct {
	store *Store
	clock Clock
	key   string
}

func (g GetCommand) Run() string {
	result, ok := g.store.Get(g.key)
	if !ok {
		return "$-1\r\n"
	}

	expiryTime := result.ExpiryTime()
	if expiryTime != nil && g.clock.Now().After(*expiryTime) {
		return "$-1\r\n"
	}

	return bulkString(result.Data())
}

type InfoKind string

const (
	InfoKindReplication InfoKind = "replication"
)

func NewInfoCommand(config *Config, infoKind InfoKind) *InfoCommand {
	return &InfoCommand{
		config:   config,
		infoKind: infoKind,
	}
}

type InfoCommand struct {
	config   *Config
	infoKind InfoKind
}

func (i *InfoCommand) Run() string {
	return bulkString("role:" + string(i.config.Replication.Role.String()))
}

type PingCommand struct{}

func (p PingCommand) Run() string {
	return simpleString("PONG")
}

func NewSetCommand(
	store *Store,
	key,
	value string,
	options ...func(*SetCommand),
) *SetCommand {
	result := &SetCommand{
		store: store,
		key:   key,
		value: value,
	}
	for _, option := range options {
		option(result)
	}
	return result
}

type SetCommand struct {
	store      *Store
	key        string
	value      string
	expiryTime *time.Time
}

func (s *SetCommand) Run() string {
	if s.expiryTime == nil {
		s.store.Set(s.key, s.value)
	} else {
		s.store.SetWithExpiryTime(s.key, s.value, *s.expiryTime)
	}
	return simpleString("OK")
}

func (s *SetCommand) Equal(other *SetCommand) bool {
	return reflect.DeepEqual(s.store, other.store) &&
		s.key == other.key &&
		s.value == other.value &&
		s.expiryTimesEqual(other)
}

func (s *SetCommand) expiryTimesEqual(other *SetCommand) bool {
	var expiryTimesEqual bool
	if s.expiryTime == nil && other.expiryTime == nil {
		expiryTimesEqual = true
	} else if s.expiryTime == nil || other.expiryTime == nil {
		expiryTimesEqual = false
	} else {
		expiryTimesEqual = s.expiryTime.Equal(*other.expiryTime)
	}
	return expiryTimesEqual
}

func ExpiryTime(t time.Time) func(*SetCommand) {
	return func(command *SetCommand) {
		command.expiryTime = &t
	}
}

func bulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func simpleString(s string) string {
	return fmt.Sprintf("+%s\r\n", s)
}

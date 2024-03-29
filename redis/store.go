package redis

func NewStore() *Store {
	return &Store{
		keyValues: make(map[string]string),
	}
}

type Store struct {
	keyValues map[string]string
}

func (s *Store) Get(key string) (result string, ok bool) {
	result, ok = s.keyValues[key]
	return
}

func (s *Store) Set(key, value string) {
	s.keyValues[key] = value
}

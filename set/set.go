package set

import (
	"sync"
)

type Set struct {
	lock sync.RWMutex
	set  map[string]bool
}

func New() *Set {
	return &Set{set: map[string]bool{}}
}

func (s *Set) Add(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.set[key] = true
}

func (s *Set) Has(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.set[key]
	return ok
}

func (s *Set) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.set, key)
}

// AddNX 如果集合没有该string则添加
// return true:添加  false:添加失败
func (s *Set) AddNX(key string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.set[key]
	if ok {
		return false
	}
	s.set[key] = true
	return true

}

func (s *Set) Len() int {
	return len(s.set)
}

package dict

import "sync"

type SyncDict struct {
	m    sync.Map
	lock sync.Mutex
}

func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

func (s *SyncDict) Get(key string) (val interface{}, exist bool) {
	if value, ok := s.m.Load(key); ok {
		return value, true
	}
	return nil, false
}

func (s *SyncDict) Len() int {
	count := 0
	// 在进行range操作的时候，syncMap已经是线程安全的了，此时不需要进行加锁操作
	s.m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (s *SyncDict) Put(key string, value interface{}) (result int) {
	_, ok := s.m.Load(key)
	s.m.Store(key, value)
	if ok {
		return 0
	}
	return 1
}

func (s *SyncDict) PutIfAbsent(key string, value interface{}) (result int) {
	_, loaded := s.m.LoadOrStore(key, value)
	if loaded {
		return 0
	}
	return 1
}

func (s *SyncDict) PutIfExists(key string, value interface{}) (result int) {
	// CAS锁尝试插入，自旋操作
	for {
		load, ok := s.m.Load(key)
		if !ok {
			return 0
		}
		// 乐观锁尝试插入
		if s.m.CompareAndSwap(key, load, value) {
			return 1
		}
	}
}

func (s *SyncDict) Remove(key string) (result int) {
	_, loaded := s.m.LoadAndDelete(key)
	if !loaded {
		return 0
	}
	return 1
}

func (s *SyncDict) ForEach(consumer Consumer) {
	s.m.Range(func(key, value interface{}) bool {
		consumer(key.(string), value)
		return true
	})
}

func (s *SyncDict) Keys() []string {
	keys := make([]string, 0)
	s.m.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

func (s *SyncDict) RandomKeys(n int) (keys []string) {
	// 每次随机返回一个，返回列表可重复
	keys = make([]string, n)
	for i := 0; i < n; i++ {
		s.m.Range(func(key, value interface{}) bool {
			keys = append(keys, key.(string))
			return false
		})
	}
	return keys
}

func (s *SyncDict) RandomDistinctKeys(n int) (keys []string) {
	keys = make([]string, n)
	// 尝试最大可能的返回，如果字典不够的话进行截断返回
	s.m.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return len(keys) < n
	})
	return keys
}

func (s *SyncDict) Clear() {
	// 直接重新赋值
	*s = *MakeSyncDict()
}

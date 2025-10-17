package atomic

import "sync/atomic"

/**
 * 使用 atomic包实现原子化bool类型
 */

type Boolean uint32

func (b *Boolean) Get() bool {
	return atomic.LoadUint32((*uint32)(b)) == 1
}

func (b *Boolean) Set(value bool) {
	if value {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}

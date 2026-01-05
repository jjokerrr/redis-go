package dict

type Consumer func(key string, value interface{}) bool

// Dict 字典存储基础数据结构接口
type Dict interface {
	Get(key string) (val interface{}, exist bool)
	Len() int
	Put(ket string, value interface{}) (result int)
	PutIfAbsent(key string, value interface{}) (result int)
	PutIfExists(key string, value interface{}) (result int)
	Remove(key string) (result int)
	ForEach(consumer Consumer)                // 迭代方法，入参传入消费方法
	Keys() []string                           // 返回当前存储的所有key
	RandomKeys(n int) (keys []string)         // 返回随机的n个key，可存在重复
	RandomDistinctKeys(n int) (keys []string) // 返回随机的n个key
	Clear()
}

package resp

type Reply interface {
	ToBytes() []byte // 回复消息的通用接口，只要符合resp协议的所有内容都可以继承这个接口
}

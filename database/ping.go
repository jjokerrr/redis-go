package database

import (
	"redis-go/interface/resp"
	"redis-go/resp/reply"
)

func PingFunc(database *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}

// init函数，这个函数会在包加载的时候自动执行
func init() {
	RegisterCommand("PING", PingFunc, 1)
}

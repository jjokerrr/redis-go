package database

import (
	"redis-go/interface/database"
	"redis-go/interface/resp"
	"redis-go/lib/utils"
	"redis-go/resp/reply"
)

// strings 命令实现，包含get，set，setnx，setex，strlen等命令

func execGet(db *DB, args [][]byte) resp.Reply {
	val, exist := db.GetEntity(string(args[0]))
	if !exist {
		return reply.MakeNullBulkReply()
	}
	return reply.MakeBulkReply(val.Data.([]byte))

}

// set
func execSet(db *DB, args [][]byte) resp.Reply {
	db.PutEntity(string(args[0]), &database.DataEntity{Data: args[1]})
	db.addAof(utils.ToCmdLineWithName("SET", args...)) // 写命令应当发送, ...代表拆包操作
	return reply.MakeOKReply()
}

// setnx
func execSetNX(db *DB, args [][]byte) resp.Reply {
	exists := db.PutIfAbsent(string(args[0]), &database.DataEntity{Data: args[1]})
	return reply.MakeIntReply(int64(exists))
}

// setex
func execSetEX(db *DB, args [][]byte) resp.Reply {
	exists := db.PutIfExists(string(args[0]), &database.DataEntity{Data: args[1]})
	return reply.MakeIntReply(int64(exists))
}

// getset,设置键值对并且返回旧值
func execGetSet(db *DB, args [][]byte) resp.Reply {
	entity, ok := db.GetEntity(string(args[0]))
	db.PutEntity(string(args[0]), &database.DataEntity{Data: args[1]})
	if ok {
		return reply.MakeBulkReply(entity.Data.([]byte))
	}
	return reply.MakeNullBulkReply()
}

// strlen
func execStrLen(db *DB, args [][]byte) resp.Reply {
	entity, ok := db.GetEntity(string(args[0]))
	if !ok {
		return reply.MakeIntReply(0)
	}
	return reply.MakeIntReply(int64(len(entity.Data.([]byte))))
}

func init() {
	RegisterCommand("get", execGet, 1)
	RegisterCommand("set", execSet, 2)
	RegisterCommand("setnx", execSetNX, 2)
	RegisterCommand("setex", execSetEX, 2)
	RegisterCommand("getset", execGetSet, 2)
	RegisterCommand("strlen", execStrLen, 1)
}

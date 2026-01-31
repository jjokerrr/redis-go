package database

import (
	"redis-go/interface/resp"
	"redis-go/lib/utils"
	"redis-go/lib/wildcard"
	"redis-go/resp/reply"
)

// 实现redis的常见命令

// 删除命令，返回值是对应删除个数
func execDel(db *DB, args [][]byte) resp.Reply {
	if len(args) == 0 {
		return reply.MakeIntReply(0)
	}
	keys := make([]string, 0, len(args))
	for _, arg := range args {
		keys = append(keys, string(arg))
	}
	db.addAof(utils.ToCmdLineWithName("DEL", args...))
	return reply.MakeIntReply(int64(db.Removes(keys...)))
}

// 判断key是否存在,返回存在的键个数
func execExists(db *DB, args [][]byte) resp.Reply {
	if len(args) == 0 {
		return reply.MakeIntReply(0)
	}
	exists := 0
	for _, arg := range args {
		_, exist := db.GetEntity(string(arg))
		if exist {
			exists++
		}
	}
	return reply.MakeIntReply(int64(exists))
}

// 刷新db
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	db.addAof(utils.ToCmdLineWithName("FLUSH", args...))
	return reply.MakeOKReply()
}

// type 返回键对应的数据类型，这里暂时只返回string todo：完善这个
func execType(db *DB, args [][]byte) resp.Reply {
	entity, b := db.GetEntity(string(args[0]))
	if !b {
		return reply.MakeStatusReply("none")
	}
	if _, ok := entity.Data.(string); ok {
		return reply.MakeStatusReply("string")
	}
	return reply.MakeUnknownReply()
}

// rename 将键进行重命名操作
func execRename(db *DB, args [][]byte) resp.Reply {
	// 检查键是否存在
	src := string(args[0])
	dst := string(args[1])
	// 检查原始键是否存在
	srcEntity, srcExist := db.GetEntity(src)
	if !srcExist {
		return reply.MakeStandardErrorReply("no such key")
	}
	db.PutEntity(dst, &srcEntity)
	db.Remove(src)
	db.addAof(utils.ToCmdLineWithName("RENAME", args...))
	return reply.MakeOKReply()
}

// rename 增强版，如果key2存在，拒绝操作
func execRenameNx(db *DB, args [][]byte) resp.Reply {
	dst := string(args[1])
	//  检查目标键是否存在
	_, dstExist := db.GetEntity(dst)
	db.addAof(utils.ToCmdLineWithName("RENAMENX", args...))
	if dstExist {
		return reply.MakeStandardErrorReply("key already exists")
	}
	return execRename(db, args)
}

// keys 遍历全部的键，返回键集合, 实现统配匹配
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0) // Store all matching keys
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

func init() {
	RegisterCommand("del", execDel, -1)
	RegisterCommand("exists", execExists, -1)
	RegisterCommand("flush", execFlushDB, 0)
	RegisterCommand("type", execType, 1)
	RegisterCommand("rename", execRename, 2)
	RegisterCommand("renamenx", execRenameNx, 2)
	RegisterCommand("keys", execKeys, 1)
}

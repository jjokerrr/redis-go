package database

import (
	"redis-go/datastruct/dict"
	"redis-go/interface/database"
	"redis-go/interface/resp"
	"redis-go/resp/reply"
	"strings"
)

// DB 是最基础数据执行单元，对应redis中的16个数据库。
// 这里不要和Database接口的作用混淆
// Database 对应的级别是数据库模式，是单机数据库还是集群数据库
// 这里简单的给出这里的分层结构
// DataBase ----  最外层数据库，单机OR集群
// DB ---- 数据库实例，一个redis中存在16个数据库，以实例存在，屏蔽底层实现
// dict ---- 具体的数据存储数据结构
type DB struct {
	index int
	data  dict.Dict
}

func MakeDB() *DB {
	return &DB{
		index: 0,
		data:  dict.MakeSyncDict(),
	}
}

// ExecFunc 执行方法，针对db实例级别
type ExecFunc func(db *DB, args [][]byte) resp.Reply

// CommandLine 命令行结构体对应命令行参数的，增强语义特性
type CommandLine [][]byte

// Exec 命令执行方法，命令执行的入口
func (db *DB) Exec(client resp.Connection, cmdLine CommandLine) resp.Reply {
	// 1. 获取命令
	cmdName := strings.ToLower(string(cmdLine[0]))
	// 2. 获取命令元信息
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeStandardErrorReply("[Command Error] Unknow command + " + cmdName)
	}
	// 3. 进行参数检查，因为存在可变参数的情况，这里抽象一下
	if !ValidateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	// 4. 命令执行
	return cmd.exec(db, cmdLine[1:])
}

func ValidateArity(arity int, args [][]byte) bool {
	if arity >= 0 {
		return arity == len(args)
	}
	// 当参数为负数的时候，说明这是一个可变参数,参数数量要大于绝对值的数量
	return -arity <= len(args)
}

// GetEntity 获取指定key的数据实体
func (db *DB) GetEntity(key string) (database.DataEntity, bool) {
	// 从底层数据
	val, exist := db.data.Get(key)
	if !exist {
		return database.DataEntity{}, false
	}
	return database.DataEntity{Data: val}, true

}

// PutEntity 写入实体
func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity.Data)
}

// PutIfExists 存在则更新
func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity.Data)
}

// PutIfAbsent 存在则放弃写入
func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity.Data)
}

// Remove 删除数据
func (db *DB) Remove(key string) int {
	return db.data.Remove(key)
}

// Removes 批量删除, 返回值记录成功删除个数，作为fail safe逻辑
func (db *DB) Removes(keys ...string) int {
	deleted := 0 //
	for _, key := range keys {
		result := db.data.Remove(key)
		if result > 0 {
			deleted++
		}
	}
	return deleted
}

// Flush 清空数据库
func (db *DB) Flush() {
	db.data.Clear()
}

//下面简单写一个选项模式的内容，主要是联系使用，对于本文的借口没有实际意义

type DBOption struct { // 接受策略的对象
	index int
	data  dict.Dict
}

type Option interface { // 策略借口
	apply(option *DBOption)
}

type FuncOption struct { // 策略实现
	f func(option *DBOption)
}

func (fo *FuncOption) apply(option *DBOption) {
	fo.f(option)
}

func WithIndex(index int) Option {
	return &FuncOption{
		f: func(option *DBOption) {
			option.index = index
		},
	}
}

func WithData(data dict.Dict) Option {
	return &FuncOption{
		f: func(option *DBOption) {
			option.data = data
		},
	}
}

const (
	defaultDBIndex = 0
)

func NewDB(opts ...Option) *DB {
	option := &DBOption{
		index: defaultDBIndex,
		data:  dict.MakeSyncDict(),
	}
	for _, opt := range opts {
		opt.apply(option)
	}
	return &DB{
		index: option.index,
		data:  option.data,
	}
}

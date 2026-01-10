package database

import (
	"redis-go/config"
	"redis-go/interface/resp"
	"redis-go/lib/logger"
	"redis-go/resp/reply"
	"strconv"
)

// 单体模式数据库
type StandaloneDatabase struct {
	dbSet []*DB
}

func NewStandaloneDatabase() *StandaloneDatabase {
	// 创建一个数据库实例
	database := &StandaloneDatabase{}
	if config.Properties.Databases <= 0 {
		config.Properties.Databases = 16
	}
	database.dbSet = make([]*DB, config.Properties.Databases)
	for i := 0; i < config.Properties.Databases; i++ {
		db := NewDB(WithIndex(i)) // 设置带编号的数据库
		database.dbSet[i] = db
	}
	return database
}

func (s *StandaloneDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			// 错误处理
			logger.Error("error occurs when processing command, %v", err)
		}
	}()
	// 拦截检验当前是否选择db命令
	if len(args) > 0 && string(args[0]) == "select" {
		if len(args) != 2 {
			return reply.MakeArgNumErrReply("[Database exec error] select database args error")
		}
		return execSelect(client, s, args[1:])
	}
	return s.dbSet[client.GetDBIndex()].Exec(client, args)
}

// execSelect sets the current database for the client connection.
// select x
func execSelect(c resp.Connection, database *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid DB index")
	}
	if dbIndex < 0 || dbIndex >= len(database.dbSet) {
		return reply.MakeStandardErrorReply("ERR DB index out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeIntReply(int64(dbIndex))
}

func (s *StandaloneDatabase) AfterClientClose(c resp.Connection) {
	logger.Info("client closed ... ")
}

func (s *StandaloneDatabase) Close() {
	logger.Info("database closed ... ")
}

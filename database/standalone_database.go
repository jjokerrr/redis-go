package database

import (
	"redis-go/aof"
	"redis-go/config"
	"redis-go/constant"
	"redis-go/interface/resp"
	"redis-go/lib/logger"
	"redis-go/resp/reply"
	"strconv"
	"strings"
)

// 单体模式数据库
type StandaloneDatabase struct {
	dbSet      []*DB
	aofHandler *aof.AofHandler
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
	// 数据库创建完成，进行初始化操作，加载持久化文件
	if config.Properties.AppendOnly {
		handler, err := aof.NewAofHandler(database)
		database.aofHandler = handler
		if err != nil {
			logger.Error("failed to open aof file, %v", err)
			panic("fatal error")
		}
		// 由于匿名函数使用了外部变量，这里参数指向了外部变量的地址，所有的匿名方法都指向了同一个值
		for _, db := range database.dbSet {
			sdb := db
			db.addAof = func(line constant.CommandLine) {
				logger.Info("[database exec] add aof, current command: ", line)
				database.aofHandler.AddHandler(sdb.index, line)
			}
		}
	}
	return database
}

func (s *StandaloneDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			// 错误处理
			logger.Error("error occurs when processing command", err)
		}
	}()
	commandName := string(args[0])
	logger.Info("[database exec] current command: ", args)
	// 拦截检验当前是否选择db命令
	if len(args) > 0 && strings.ToLower(commandName) == "select" {
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

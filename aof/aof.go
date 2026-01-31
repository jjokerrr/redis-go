package aof

import (
	"io"
	"os"
	"redis-go/config"
	"redis-go/constant"
	databaseface "redis-go/interface/database"
	"redis-go/lib/logger"
	"redis-go/lib/utils"
	"redis-go/resp/connection"
	"redis-go/resp/parser"
	"redis-go/resp/reply"
	"slices"
	"strconv"
)

const aofBufferSize = 1 << 16

// aof通过异步方式传递命令数据，命令在执行结尾将数据通过payload传递到管道中
// 显然由于database作为命令的实际执行者，我们是可以获取到当前命令执行所对应的dbIndex
// 当然存在一个特例。当我们的select命令是额外执行的，所以我们要增加一个dbIndex命令，防止select命令丢失导致数据不一致问题
type payload struct {
	cmd     constant.CommandLine
	dbIndex int
}

type AofHandler struct {
	db          databaseface.Database // db对象
	aofChan     chan *payload         // 命令管道，接收命令进行持久化
	aofFile     *os.File              // 命令持久化文件
	aofFileName string                // 命令持久化文件
	currDB      int                   // 当前持久化所在db，和payload中的db对照使用
}

func NewAofHandler(db databaseface.Database) (*AofHandler, error) {
	handler := &AofHandler{
		db: db,
	}
	handler.aofFileName = config.Properties.AppendFilename
	// 加载持久化文件
	handler.LoadAof()
	file, err := os.OpenFile(handler.aofFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = file
	// 创建持久化管道
	handler.aofChan = make(chan *payload, aofBufferSize)
	// 异步处理命令
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

func (handler *AofHandler) Close() error {
	return handler.aofFile.Close()
}

func (handler *AofHandler) LoadAof() {
	// 1. 加载文件
	logger.Info("[load aof file] start load aof file")
	file, err := os.Open(handler.aofFileName)
	if err != nil {
		logger.Info("[load aof file] the aof file is not exist or open error", err)
		return
	}
	// 2. 存在文件才会继续加载,利用parseStream进行文件的逐行读取
	defer file.Close()
	ch := parser.ParseStream(file)
	// 3. 通过fakeConn来进行命令执行装载
	fakeConn := &connection.Connection{}
	for payload := range ch {
		if payload.Err != nil {
			// If the error is EOF or unexpected EOF, break the loop
			if payload.Err == io.EOF || payload.Err == io.ErrUnexpectedEOF {
				// End of file
				break
			}
			logger.Error("[load aof file] parse error", payload.Err)
			continue
		}
		if payload.Data == nil {
			continue
		}
		// 4. 获取命令
		bulkReply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			continue
		}
		// 为了保证回放的命令不再二次写入aof文件中，采用提前初始化AddAof方法的方式将其转换为空方法
		rep := handler.db.Exec(fakeConn, bulkReply.Args) // 执行命令写入
		if reply.IsErrReply(rep) {
			logger.Error("Execute AOF command error")
		}
	}
	logger.Info("[load aof file] load aof file success")
}

func (handler *AofHandler) handleAof() {
	handler.currDB = 0
	// 从ch中获取命令，将命令持久化到文件中
	for pl := range handler.aofChan {
		logger.Info("[handle aof] write command to file: ", pl.cmd)
		var cmdToWrite []byte
		if pl.dbIndex != handler.currDB {
			// 出现db切换现象，进行补充select命令
			selectCmd := reply.MakeMultiBulkReply(utils.ToCmdLine("select", strconv.Itoa(pl.dbIndex))).ToBytes()
			cmdToWrite = slices.Concat(cmdToWrite, selectCmd)
		}
		cmd := reply.MakeMultiBulkReply(pl.cmd).ToBytes()
		cmdToWrite = slices.Concat(cmdToWrite, cmd)
		// 将命令持久化的文件中
		_, err := handler.aofFile.Write(cmdToWrite)
		if err != nil {
			logger.Error("[handle aof error] write cmd to file err! current command: " + string(cmd))
			continue
		}
		// 将命令刷盘，防止由于内存中的命令尚未持久化导致数据丢失
		handler.aofFile.Sync()

	}
}

func (handler *AofHandler) AddHandler(index int, line constant.CommandLine) {
	// 合法性校验
	if handler.aofChan == nil || !config.Properties.AppendOnly {
		handler.aofChan = make(chan *payload, 100)
	}
	// 写入到channel中
	handler.aofChan <- &payload{
		cmd:     line,
		dbIndex: index,
	}
}

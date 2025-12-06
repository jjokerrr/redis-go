package handler

import (
	"context"
	"io"
	"net"
	"redis-go/database"
	databaseface "redis-go/interface/database"
	"redis-go/lib/logger"
	"redis-go/lib/sync/atomic"
	"redis-go/resp/connection"
	"redis-go/resp/parser"
	"redis-go/resp/reply"
	"sync"
)

type RespHandler struct {
	activeConn sync.Map // 存放简历链接的Connection对象
	db         databaseface.Database
	closing    atomic.Boolean
}

func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) error {
	if h.closing.Get() { // 当前处于关闭状态，拒绝后续的client链接
		// 关闭当前新的链接
		_ = conn.Close()
	}
	// 创建客户端链接
	client := connection.NewConnection(conn)
	h.activeConn.Store(client, struct{}{})

	// 流式的接收消息
	payLoads := parser.ParseStream(conn)
	for payload := range payLoads {
		// 错误处理，直接进行打印操作
		if payload.Err != nil {
			// 出现问题需要主动的去关闭链接
			if payload.Err == io.EOF {
				// 主动关闭当前链接
				h.closeClient(client)
				continue
			} else {
				_ = client.Write(reply.MakeStandardErrorReply(payload.Err.Error()).ToBytes())
				continue
			}
		}
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		// 尝试执行命令并返回执行结果
		// 简单构建一个echo返回结果
		bulkReply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require bulk reply")
			continue
		}
		res := h.db.Exec(client, bulkReply.Args)
		err := client.Write(res.ToBytes())
		if err != nil {
			return err
		}

	}
	return nil
}

func (h *RespHandler) Close() error {
	h.closing.Set(true)
	h.activeConn.Range(func(key, value interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	h.db.Close()
	return nil
}

func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

func MakeHandler() *RespHandler {
	db := database.NewEchoDatabase()
	return &RespHandler{
		db: db,
	}
}

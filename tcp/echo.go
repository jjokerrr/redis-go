package tcp

import (
	"bufio"
	"context"
	"io"
	"net"
	"redis-go/lib/logger"
	"redis-go/lib/sync/atomic"
	"redis-go/lib/wait"
	"sync"
	"time"
)

type EchoHandler struct {
	activateConn sync.Map       // 活跃链接
	closing      atomic.Boolean // server状态
}

type EchoClient struct {
	Conn net.Conn
	Wait wait.Wait // 记录处理过程
}

func (e *EchoClient) Close() error {
	// 先进行优雅退出
	e.Wait.WaitWithTimeout(5 * time.Second)
	// 超时未退出，进行手动关系
	err := e.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (e *EchoHandler) Handle(ctx context.Context, conn net.Conn) error {
	if e.closing.Get() { // 当前处于关闭状态，拒绝后续的client链接
		// 关闭当前新的链接
		_ = conn.Close()
	}
	// 创建客户端链接
	client := &EchoClient{
		Conn: conn,
	}
	e.activateConn.Store(client, struct{}{})
	// 循环发送和接收消息
	reader := bufio.NewReader(client.Conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 处理结束，尝试关闭链接
				logger.Info("client closed")
				e.activateConn.Delete(client)
				_ = client.Close()
			} else {
				logger.Error(err)
				e.activateConn.Delete(client)
			}
			return nil
		}
		client.Wait.Add(1)
		// 写回去
		_, err = conn.Write([]byte(line))
		if err != nil {
			return err
		}
		client.Wait.Done()
	}

}

func (e *EchoHandler) Close() error {
	logger.Info("server closing....")
	// 先改变自身状态，拒绝后续的client链接
	e.closing.Set(true)
	// 进行其他进行的优雅退出
	e.activateConn.Range(func(key, value interface{}) bool {
		client := key.(*EchoClient)
		// 等待客户端自动退出, 这里不能一场抛出，避免阻塞后续的关闭流程
		_ = client.Close()
		return true
	})
	return nil
}

func MakeHandler() *EchoHandler {
	return &EchoHandler{}
}

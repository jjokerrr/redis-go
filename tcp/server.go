package tcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"redis-go/interface/tcp"
	"redis-go/lib/logger"
	"sync"
	"syscall"
)

type Config struct {
	Address string
}

// ListenAndServeWithSignal 绑定端口，注册新号
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	go func() {
		<-closeChan
		_ = listener.Close()
		_ = handler.Close()
	}()

	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background() // 创建一个空白的context
	wg := sync.WaitGroup{}      // 出现错误链接的时候，进行优雅退出
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(err)
			break
		}
		logger.Info(fmt.Sprintf("get new connection: %s", conn.RemoteAddr().String()))
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = handler.Handle(ctx, conn)
		}()
	}
	wg.Wait()
}

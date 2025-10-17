package tcp

import (
	"context"
	"net"
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn) error // server处理逻辑，ctx接收终端信号等
	Close() error                                    // server需实现优雅退出
}

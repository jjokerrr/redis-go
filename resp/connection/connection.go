package connection

import (
	"net"
	"redis-go/lib/wait"
	"sync"
	"time"
)

// Connection 表示客户端和服务端的连接
type Connection struct {
	conn         net.Conn   // 底层的网络连接
	waitingReply wait.Wait  // 等待完成响应的同步器
	mu           sync.Mutex // 发送响应时的互斥锁
	selectedDB   int        // 选择的数据库的编号
}

func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

func (c *Connection) SelectDB(i int) {
	c.selectedDB = i
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{conn: conn}
}

func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	err := c.conn.Close()
	return err
}

func (c *Connection) getRemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Write 向conn中进行写数据的方法，放置在并发环境造成写的问题，这里增加互斥锁保证写的串行化
func (c *Connection) Write(bytes []byte) error {
	c.mu.Lock()
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()
	c.waitingReply.Add(1)
	_, err := c.conn.Write(bytes)
	return err

}

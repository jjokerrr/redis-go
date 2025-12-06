package reply

import (
	"bytes"
	"redis-go/interface/resp"
	"strconv"
)

// ErrorReply 错误回复，实现了 Reply 的 ToBytes 方法，也实现了系统的 error 接口
// 这里使用了接口组合，将 error 接口和 Reply 接口组合在一起
type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

// BulkReply 字符串回复
type BulkReply struct {
	Arg []byte // 回复的内容，此时是不符合 RESP 协议的
}

type MultiBulkReply struct {
	Args [][]byte
}

// ToBytes 字符串回复, string 类型格式 $<len>/r/n<reply></r/n>
func (b *BulkReply) ToBytes() []byte {
	if len(b.Arg) == 0 {
		return MakeEmptyBulkReply().ToBytes()
	}
	return []byte("$" + strconv.Itoa(len(b.Arg)) + CRLF + string(b.Arg) + CRLF)
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

func (m *MultiBulkReply) ToBytes() []byte {
	if len(m.Args) == 0 {
		return MakeEmptyMultiBulkReply().ToBytes()
	}

	buffer := bytes.Buffer{}
	buffer.WriteString("*" + strconv.Itoa(len(m.Args)) + CRLF)
	for _, arg := range m.Args {
		buffer.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
	}
	return buffer.Bytes()
}

func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

type StandardErrorReply struct {
	Status string
}

func (s *StandardErrorReply) ToBytes() []byte {
	return []byte("-" + s.Status + CRLF)
}

func MakeStandardErrorReply(status string) *StandardErrorReply {
	return &StandardErrorReply{
		Status: status,
	}
}

type IntReply struct {
	Code int64
}

func (i *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(i.Code, 10) + CRLF)
}

func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

type StatusReply struct {
	Status string
}

func (s *StatusReply) ToBytes() []byte {
	return []byte("+" + s.Status + CRLF)
}

func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

func IsErrReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}

package parser

import (
	"bufio"
	"errors"
	"io"
	"redis-go/interface/resp"
	"redis-go/lib/logger"
	"redis-go/resp/reply"
	"runtime/debug"
	"strconv"
)

type PayLoad struct {
	Data resp.Reply
	Err  error
}

type readState struct {
	readingMultiLine  bool     // 是否正在读取多行数据
	expectedArgsCount int      // 期望的参数数量
	msgType           byte     // 消息类型
	args              [][]byte // 参数
	bulkLen           int64    // Bulk 回复的长度,
}

func (state *readState) isDone() bool {
	return state.expectedArgsCount > 0 && state.expectedArgsCount <= len(state.args)
}

// ParseStream 流式接收命令，解析命令
func ParseStream(reader io.Reader) <-chan *PayLoad {
	ch := make(chan *PayLoad)
	go parse0(reader, ch)
	return ch
}

func parse0(reader io.Reader, ch chan *PayLoad) {
	defer func() {
		if err := recover(); err != nil { // recover 从panic中恢复获取错误原因
			logger.Error(string(debug.Stack()))
		}
	}()
	// 解析命令，并将命令发送到ch
	bufReader := bufio.NewReader(reader)
	var state = readState{}

	for {
		// 不带分隔符的单条命令
		line, ioErr, err := readLine(bufReader, &state)
		logger.Info("parse line: ", string(line))
		// 非io一场，那么可以继续解析
		if err != nil {
			if ioErr { // io异常，说明出现了丢包问题或者是其他的解析问题，或者是链接已经断开，会影响后续的命令解析，这里直接断开链接
				logger.Info("[parse0 error]: io error")
				ch <- &PayLoad{
					Err: err,
				}
				close(ch)
				return
			}
			logger.Info("[parse0 error]: parse error" + err.Error())
			ch <- &PayLoad{
				Err: err,
			}
			state = readState{} // 重制状态
			continue
		}

		// 处理空行情况
		if len(line) == 0 {
			continue
		}

		// 单行命令
		if !state.readingMultiLine {
			switch line[0] {
			case '$':
				err := parseBulkHeader(line, &state)
				if err != nil {
					ch <- &PayLoad{
						Err: err,
					}
					state = readState{}
					continue
				}
				if state.bulkLen < 0 {
					ch <- &PayLoad{
						Data: reply.MakeNullBulkReply(),
					}
					state = readState{}
					continue
				}
			case '*':
				err := parseMultiBulkHeader(line, &state)
				if err != nil {
					ch <- &PayLoad{
						Err: err,
					}
					state = readState{}
					return
				}
				if state.expectedArgsCount == 0 { // 空数组，直接提交命令
					ch <- &PayLoad{Data: reply.MakeEmptyMultiBulkReply()}
					state = readState{}
					continue
				}

			default: // 完成单行命令的处理
				singleLine, err := parseSingleLine(line, &state)
				if err != nil {
					ch <- &PayLoad{
						Err: err,
					}
					state = readState{}
				}
				ch <- &PayLoad{
					Data: singleLine,
				}
				state = readState{}
				continue

			}
		} else {
			err = readBody(line, &state)
			if err != nil {
				ch <- &PayLoad{
					Err: err,
				}
				state = readState{}
				continue
			}
			if state.isDone() { // 完成之后发送完整命令
				var res resp.Reply
				if state.msgType == '*' {
					res = reply.MakeMultiBulkReply(state.args)
				} else {
					res = reply.MakeBulkReply(state.args[0])
				}
				ch <- &PayLoad{
					Data: res,
				}
				state = readState{}
			}
		}
	}

}

func readBody(line []byte, r *readState) error {
	// 多行字符串和数组类型处理
	if !r.readingMultiLine {
		return errors.New("[parseBody error]: parse body error")
	}

	var err error
	if line[0] == '$' {
		r.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("[parseBody error]: parse body error")
		}
		// 兼容零值情况
		if r.bulkLen == 0 {
			r.args = append(r.args, []byte{})
			r.bulkLen = 0 // 兼容readLine
		}
	} else {
		r.args = append(r.args, line)
	}
	return nil
}

func parseSingleLine(line []byte, r *readState) (resp.Reply, error) {
	r.msgType = line[0]
	switch r.msgType {
	case '+':
		return reply.MakeStatusReply(string(line[1:])), nil
	case '-':
		return reply.MakeStandardErrorReply(string(line[1:])), nil
	case ':':
		parseInt, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, err
		}
		return reply.MakeIntReply(parseInt), nil
	}
	return nil, errors.New("[parseSingleLine] parse single line error")
}

func parseBulkHeader(line []byte, r *readState) error {
	// 多行字符串头部解析
	var err error
	r.msgType = line[0]
	r.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		return errors.New("[parse BulkHeader] parse bulk header error")
	}
	if r.bulkLen < 0 {
		r.bulkLen = -1
		return nil
	} else {
		r.readingMultiLine = true
		r.expectedArgsCount = 1

		r.args = make([][]byte, 0, 1)
	}
	return nil

}

func parseMultiBulkHeader(line []byte, r *readState) error {
	var err error
	r.msgType = line[0]
	expectedArgsCount, err := strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		return err
	}
	if expectedArgsCount < 0 {
		return errors.New("[parse MultiBulkHeader] parse multi bulk header error")
	} else {
		r.readingMultiLine = true
		r.expectedArgsCount = int(expectedArgsCount)
		r.args = make([][]byte, 0, expectedArgsCount)
	}
	return nil
}

// readLine 读取一行数据，读取方式有两种，通过readState进行区分。
// 对于多行字符串，首先要按照读取字符串头，其中包含字符串的一部分元信息，这样下一次读取的时候直接按照元信息读取对应长度
func readLine(reader *bufio.Reader, r *readState) ([]byte, bool, error) {
	if r.bulkLen == 0 {
		// read a normal line.
		// when client closing, there will be an io.EOF error
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(line) < 2 || line[len(line)-2] != '\r' || line[len(line)-1] != '\n' {
			return nil, false, errors.New("[readLine error]: line is not a resp Protocol")
		}
		return line[:len(line)-2], false, nil
	} else {
		line := make([]byte, r.bulkLen+2)
		_, err := io.ReadFull(reader, line)
		if err != nil {
			return nil, true, err
		}
		if len(line) < 2 || line[len(line)-2] != '\r' || line[len(line)-1] != '\n' {
			return nil, false, errors.New("[readLine error]: line is not a resp Protocol")
		}
		r.bulkLen = 0 // 读完了，把多行读取清零
		return line[:len(line)-2], false, nil
	}

}

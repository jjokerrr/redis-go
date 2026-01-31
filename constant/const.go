package constant

import "strings"

type CommandLine [][]byte

// CommandLine 转换成字符串
func (c CommandLine) String() string {
	if len(c) == 0 {
		return ""
	}

	// 使用 strings.Builder 减少内存分配开销
	var builder strings.Builder
	for i, arg := range c {
		builder.Write(arg)
		// 在参数之间添加空格
		if i < len(c)-1 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

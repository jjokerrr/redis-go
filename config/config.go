package config

import (
	"bufio"
	"errors"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type ServerProperties struct {
	Bind           string   `cfg:"bind"`
	Port           int      `cfg:"port"`
	AppendOnly     bool     `cfg:"appendOnly"`
	AppendFilename string   `cfg:"appendFilename"`
	MaxClients     int      `cfg:"maxClients"`
	RequirePass    string   `cfg:"requirePass"`
	Databases      int      `cfg:"databases"`
	Peers          []string `cfg:"peers"`
	Self           string   `cfg:"self"`
}

var Properties *ServerProperties // 全局的配置项

func initConfig() *ServerProperties {
	return &ServerProperties{
		Bind: "0.0.0.0",
		Port: 6379,
	}
}

/*
*
加载配置文件
*/
func parseConfig(src io.Reader) *ServerProperties {
	config := &ServerProperties{}
	configMap := map[string]string{}
	// 1. 逐行读取配置文件
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		key, val, err := parseLine(line)
		if err != nil {
			continue
		}
		configMap[key] = val
	}
	configType := reflect.TypeOf(config)
	configValue := reflect.ValueOf(config)
	for i := 0; i < configType.Elem().NumField(); i++ {
		field := configType.Elem().Field(i)
		value := configValue.Elem().Field(i)
		fieldName, ok := field.Tag.Lookup("cfg")
		if !ok {
			fieldName = field.Name
		}
		val, ok := configMap[strings.ToLower(fieldName)]
		if !ok {
			continue
		}
		// 对应类型的写入和解析
		switch field.Type.Kind() {
		case reflect.Int:
			intVal, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil
			}
			value.SetInt(intVal)
		case reflect.String:
			value.SetString(val)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(val)
			if err != nil {
				return nil
			}
			value.SetBool(boolVal)
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				slice := strings.Split(val, ",")
				value.Set(reflect.ValueOf(slice))
			}
		default:
			panic("unhandled default case")
		}
	}
	return config
}

func parseLine(line string) (string, string, error) {
	if strings.HasPrefix(line, "#") {
		return "", "", errors.New("invalid config line")
	}
	configLine := strings.SplitAfter(line, " ")
	if len(configLine) != 2 {
		return "", "", errors.New("invalid config line")
	}
	return strings.ToLower(strings.TrimSpace(configLine[0])), strings.ToLower(strings.TrimSpace(configLine[1])), nil

}

func SetupConfig(configFilename string) {
	file, err := os.Open(configFilename)
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(file)
	Properties = parseConfig(file)
}

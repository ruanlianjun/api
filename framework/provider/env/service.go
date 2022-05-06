package env

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"coredemo/framework/contract"
)

type AppEnv struct {
	folder string //.env 所在的目录
	maps   map[string]string
}

func (a AppEnv) AppEnv() string {
	return a.Get("APP_ENV")
}

func (a AppEnv) IsExist(s string) bool {
	if _, ok := a.maps[s]; ok {
		return true
	}
	return false
}

func (a AppEnv) Get(s string) string {
	if val, ok := a.maps[s]; ok {
		return val
	}
	return ""
}

func (a AppEnv) All() map[string]string {
	return a.maps
}

func NewAppEnv(params ...any) (any, error) {
	if len(params) != 1 {
		return nil, errors.New("new app env err")
	}

	folder := params[0].(string)
	appEnv := AppEnv{
		folder: folder,
		//实例化环境变量，默认为开发环境
		maps: map[string]string{"APP_ENV": contract.EnvDevelopment},
	}
	file := filepath.Join(folder, ".env")

	fi, err := os.Open(file)
	if err == nil {
		defer fi.Close()
		reader := bufio.NewReader(fi)
		for {
			line, _, c := reader.ReadLine()
			if c == io.EOF {
				break
			}

			s := bytes.SplitN(line, []byte{'='}, 2)
			if len(s) < 2 {
				continue
			}
			key := string(s[0])
			val := string(s[1])
			appEnv.maps[key] = val
		}
	}

	// 获取当前的系统变量，替换 `.env` 文件里面的值
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) < 2 {
			continue
		}
		appEnv.maps[pair[0]] = pair[1]
	}
	return appEnv, nil
}

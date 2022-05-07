package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v2"

	"github.com/ruanlianjun/api/framework"
	"github.com/ruanlianjun/api/framework/contract"
)

type AppConfig struct {
	c        framework.Container //容器
	folder   string              //文件夹
	keyDelim string              //分隔符号
	lock     sync.RWMutex        //配置文件读写锁
	envMaps  map[string]string   // 所有的环境变量
	confMaps map[string]any      //配置的文件结构，key为文件名
	confRaws map[string][]byte   //配置文件的原始信息
}

func (a *AppConfig) loadConfigFile(folder string, file string) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	s := strings.Split(file, ".")
	if len(s) == 2 && (s[1] == "yaml" || s[1] == "yml") {
		name := s[0]
		bf, err := ioutil.ReadFile(filepath.Join(folder, file))
		if err != nil {
			return err
		}
		// 直接针对文本进行环境变量替换
		bf = replace(bf, a.envMaps)
		c := map[string]any{}
		if err = yaml.Unmarshal(bf, &c); err != nil {
			return err
		}
		a.confMaps[name] = c
		a.confRaws[name] = bf
		if name == "app" && a.c.IsBind(contract.AppKey) {
			if p, ok := c["path"]; ok {
				appServer := a.c.MustMake(contract.AppKey).(contract.App)
				appServer.LoadAppConfig(cast.ToStringMapString(p))
			}
		}

	}
	return nil
}

func (a *AppConfig) removeConfigFile(file string) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	s := strings.Split(file, ".")
	if len(s) == 2 && (s[1] == "yaml" || s[1] == "yml") {
		delete(a.confRaws, s[0])
		delete(a.confMaps, s[0])
	} else {
		return errors.New("remove config file: %s is not yaml or yml")
	}
	return nil
}

func NewAppConfig(params ...any) (any, error) {
	container := params[0].(framework.Container)
	envFolder := params[1].(string)
	envMaps := params[2].(map[string]string)
	a := &AppConfig{
		c:        container,
		folder:   envFolder,
		keyDelim: ".",
		lock:     sync.RWMutex{},
		envMaps:  envMaps,
		confMaps: map[string]any{},
		confRaws: map[string][]byte{},
	}

	if _, err := os.Stat(envFolder); os.IsNotExist(err) {
		return a, err
	}
	log.Printf("====================config files:%+v\n", envFolder)
	files, err := ioutil.ReadDir(envFolder)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, file := range files {
		fileName := file.Name()
		err := a.loadConfigFile(envFolder, fileName)
		if err != nil {
			log.Printf("load config file :%s not exists\n", file)
			continue
		}
	}
	//监控文件变化
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(envFolder); err != nil {
		return nil, err
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()

		for {
			select {
			case ev := <-watcher.Events:
				{
					path, _ := filepath.Abs(ev.Name)
					index := strings.LastIndex(path, string(os.PathSeparator))
					folder := path[:index]
					fileName := path[index+1:]
					if ev.Op&fsnotify.Create == fsnotify.Create {
						log.Printf("配置文件变化-创建文件： %s\n", ev.Name)
						a.loadConfigFile(folder, fileName)
					}

					if ev.Op&fsnotify.Write == fsnotify.Write {
						log.Printf("配置文件变化-写入文件： %s\n", ev.Name)
						a.loadConfigFile(folder, fileName)
					}

					if ev.Op&fsnotify.Remove == fsnotify.Remove {
						log.Printf("配置文件变化-移除文件：%s\n", ev.Name)
						a.loadConfigFile(folder, ev.Name)
					}
					if ev.Op&fsnotify.Rename == fsnotify.Rename {
						log.Printf("配置文件变化-文件名变化：%s\n", ev.Name)
						a.loadConfigFile(folder, ev.Name)
					}

				}
			case err := <-watcher.Errors:
				{
					log.Printf("监听配置文件发生错误：%s\n", err.Error())
					return
				}
			}
		}
	}()
	return a, nil
}

// replace 使用环境变量替换content中的 `env(xxx)` 环境变量
func replace(content []byte, maps map[string]string) []byte {
	if maps == nil {
		return content
	}

	for k, v := range maps {
		relKey := "env(" + k + ")"
		content = bytes.ReplaceAll(content, []byte(relKey), []byte(v))
	}
	return content
}

//查找某一个路径下面的配置
func searchMap(source map[string]any, path []string) any {
	if len(path) == 0 {
		return source
	}

	//判断下一个路径
	next, ok := source[path[0]]
	if ok {
		if len(path) == 1 {
			return next
		}

		switch next.(type) {
		case map[any]any:
			return searchMap(cast.ToStringMap(next), path[1:])
		case map[string]any:
			return searchMap(next.(map[string]any), path[1:])
		default:
			return nil
		}
	}
	return nil
}

func (a *AppConfig) find(key string) any {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return searchMap(a.confMaps, strings.Split(key, a.keyDelim))
}

// IsExist check setting is exist
func (a *AppConfig) IsExist(key string) bool {
	return a.find(key) != nil
}

// Get 获取某个配置项
func (a *AppConfig) Get(key string) interface{} {
	return a.find(key)
}

// GetBool 获取bool类型配置
func (a *AppConfig) GetBool(key string) bool {
	return cast.ToBool(a.find(key))
}

// GetInt 获取int类型配置
func (a *AppConfig) GetInt(key string) int {
	return cast.ToInt(a.find(key))
}

// GetFloat64 get float64
func (a *AppConfig) GetFloat64(key string) float64 {
	return cast.ToFloat64(a.find(key))
}

// GetTime get time type
func (a *AppConfig) GetTime(key string) time.Time {
	return cast.ToTime(a.find(key))
}

// GetString get string typen
func (a *AppConfig) GetString(key string) string {
	return cast.ToString(a.find(key))
}

// GetIntSlice get int slice type
func (a *AppConfig) GetIntSlice(key string) []int {
	return cast.ToIntSlice(a.find(key))
}

// GetStringSlice get string slice type
func (a *AppConfig) GetStringSlice(key string) []string {
	return cast.ToStringSlice(a.find(key))
}

// GetStringMap get map which key is string, value is interface
func (a *AppConfig) GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(a.find(key))
}

// GetStringMapString get map which key is string, value is string
func (a *AppConfig) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(a.find(key))
}

// GetStringMapStringSlice get map which key is string, value is string slice
func (a *AppConfig) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(a.find(key))
}

// Load a config to a struct, val should be an pointer
func (a *AppConfig) Load(key string, val interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "yaml",
		Result:  val,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(a.find(key))
}

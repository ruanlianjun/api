package app

import (
	"flag"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ruanlianjun/api/framework"
	"github.com/ruanlianjun/api/framework/utils"
)

type App struct {
	container  framework.Container
	baseFolder string // 基础路径
	appId      string // 表示当前这个app的唯一id, 可以用于分布式锁等

	configMap map[string]string // 配置加载
}

func (a App) AppID() string {
	return a.appId
}

func (a App) Version() string {
	return "0.0.1"
}

func (a App) BaseFolder() string {
	if a.baseFolder == "" {
		return a.baseFolder
	}

	return utils.GetExecDirectory()
}

func (a App) ConfigFolder() string {
	if val, ok := a.configMap["config_folder"]; ok {
		return val
	}
	return filepath.Join(a.BaseFolder(), "config")
}

func (a App) ProviderFolder() string {
	if val, ok := a.configMap["provider_folder"]; ok {
		return val
	}
	return filepath.Join(a.BaseFolder(), "app", "provider")
}

func (a App) AppFolder() string {
	if val, ok := a.configMap["app_folder"]; ok {
		return val
	}
	return filepath.Join(a.BaseFolder(), "app")
}

func (a App) LoadAppConfig(kv map[string]string) {
	for key, val := range kv {
		a.configMap[key] = val
	}
}

func NewApp(params ...any) (any, error) {
	if len(params) != 2 {
		return nil, errors.New("new app params err")
	}
	container := params[0].(framework.Container)
	baseFolder := params[1].(string)

	if baseFolder == "" {
		flag.StringVar(&baseFolder, "base_folder", "", "base_folder参数, 默认为当前路径")
		flag.Parse()
	}

	appid := uuid.New().String()
	return &App{
		container:  container,
		baseFolder: baseFolder,
		appId:      appid,
		configMap:  map[string]string{},
	}, nil
}

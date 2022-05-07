package config

import (
	"path/filepath"

	"github.com/ruanlianjun/api/framework"
	"github.com/ruanlianjun/api/framework/contract"
)

type AppConfigProvider struct{}

func (a *AppConfigProvider) Register(container framework.Container) framework.NewInstance {
	return NewAppConfig
}

func (a *AppConfigProvider) Boot(container framework.Container) error {
	return nil
}

func (a *AppConfigProvider) IsDefer() bool {
	return false
}

func (a *AppConfigProvider) Params(container framework.Container) []interface{} {
	envService := container.MustMake(contract.EnvKey).(contract.Env)
	appService := container.MustMake(contract.AppKey).(contract.App)
	env := envService.AppEnv()
	configFolder := appService.ConfigFolder()
	envFolder := filepath.Join(configFolder, env)
	return []any{container, envFolder, envService.All()}
}

func (a AppConfigProvider) Name() string {
	return contract.ConfigKey
}

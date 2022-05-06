package env

import (
	"coredemo/framework"
	"coredemo/framework/contract"
)

type AppEnvProvider struct {
	Folder string
}

func (a *AppEnvProvider) Register(container framework.Container) framework.NewInstance {
	return NewAppEnv
}

func (a *AppEnvProvider) Boot(container framework.Container) error {
	app := container.MustMake(contract.AppKey).(contract.App)
	a.Folder = app.BaseFolder()
	return nil
}

func (a *AppEnvProvider) IsDefer() bool {
	return false
}

func (a *AppEnvProvider) Params(container framework.Container) []interface{} {
	return []any{a.Folder}
}

func (a *AppEnvProvider) Name() string {
	return contract.EnvKey
}

package app

import (
	"github.com/ruanlianjun/api/framework"
	"github.com/ruanlianjun/api/framework/contract"
)

type AppProvider struct {
	BaseFolder string
}

func (a *AppProvider) Register(container framework.Container) framework.NewInstance {
	return NewApp
}

func (a *AppProvider) Boot(container framework.Container) error {
	return nil
}

func (a *AppProvider) IsDefer() bool {
	return false
}

func (a *AppProvider) Params(container framework.Container) []interface{} {
	return []any{container, a.BaseFolder}
}

func (a *AppProvider) Name() string {
	return contract.AppKey
}

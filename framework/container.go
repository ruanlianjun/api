package framework

import (
	"sync"

	"github.com/pkg/errors"
)

type Container interface {
	Bind(provider ServiceProvider) error
	IsBind(key string) bool
	Make(key string) (any, error)
	MustMake(key string) any
	MakeNew(key string, params []any) (any, error)
}

type AppContainer struct {
	Container
	//存储的服务提供者
	providers map[string]ServiceProvider
	// 存储的具体实例
	instances map[string]any
	lock      sync.RWMutex
}

func NewAppContainer() *AppContainer {
	return &AppContainer{
		providers: map[string]ServiceProvider{},
		instances: map[string]any{},
		lock:      sync.RWMutex{},
	}
}

func (a *AppContainer) PrintProviders() []string {
	var tmp []string
	for _, provider := range a.providers {
		tmp = append(tmp, provider.Name())
	}
	return tmp
}

func (a *AppContainer) Bind(provider ServiceProvider) error {
	a.lock.Lock()
	key := provider.Name()
	a.providers[key] = provider
	a.lock.Unlock()
	//如果不是延迟绑定
	if provider.IsDefer() == false {
		if err := provider.Boot(a); err != nil {
			return err
		}

		params := provider.Params(a)
		method := provider.Register(a)

		instance, err := method(params...)

		if err != nil {
			return errors.New(err.Error())
		}

		a.instances[key] = instance
	}

	return nil
}

func (a *AppContainer) IsBind(key string) bool {
	return a.findServiceProvider(key) != nil
}
func (a *AppContainer) findServiceProvider(key string) ServiceProvider {
	a.lock.RLock()
	defer a.lock.RUnlock()

	if sp, ok := a.providers[key]; ok {
		return sp
	}
	return nil
}

func (a *AppContainer) Make(key string) (any, error) {
	return a.make(key, nil, false)
}

func (a *AppContainer) MustMake(key string) any {
	server, err := a.make(key, nil, false)
	if err != nil {
		panic("container not contain key " + key)
	}
	return server
}

func (a *AppContainer) MakeNew(key string, params []any) (any, error) {
	return a.make(key, params, true)
}

func (a *AppContainer) newInstance(sp ServiceProvider, params []any) (interface{}, error) {
	if err := sp.Boot(a); err != nil {
		return nil, errors.New(err.Error())
	}

	if params == nil {
		params = sp.Params(a)
	}

	method := sp.Register(a)
	instance, err := method(params...)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

//真正实例化一个服务
func (a *AppContainer) make(key string, params []any, forceNew bool) (interface{}, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	sp := a.providers[key]
	if sp == nil {
		return nil, errors.New("contract " + key + " have not register")
	}

	if forceNew {
		return a.newInstance(sp, params)
	}

	//不需要重新初始化
	if in, ok := a.instances[key]; ok {
		return in, nil
	}

	//没有实例化
	instance, err := a.newInstance(sp, params)
	if err != nil {
		return nil, err
	}

	a.instances[key] = instance

	return instance, nil
}

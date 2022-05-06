package main

import (
	"log"

	"coredemo/framework"
	"coredemo/framework/contract"
	"coredemo/framework/provider/app"
	"coredemo/framework/provider/config"
	"coredemo/framework/provider/env"
)

func main() {

	container := framework.NewAppContainer()

	//绑定APP服务提供者
	err := container.Bind(&app.AppProvider{
		BaseFolder: "",
	})
	if err != nil {
		panic(err)
	}

	//绑定其他的提供者
	if err = container.Bind(&env.AppEnvProvider{}); err != nil {
		log.Fatal(err)
	}

	err = container.Bind(&config.AppConfigProvider{})
	if err != nil {
		panic(err)
	}

	a := container.MustMake(contract.EnvKey).(contract.Env)
	log.Printf("=========================== base_folder:%v\n", a.All())
}

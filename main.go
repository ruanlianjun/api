package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ruanlianjun/api/framework"
	"github.com/ruanlianjun/api/framework/contract"
	"github.com/ruanlianjun/api/framework/provider/app"
	"github.com/ruanlianjun/api/framework/provider/config"
	"github.com/ruanlianjun/api/framework/provider/env"
	"github.com/ruanlianjun/api/framework/provider/redis"
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

	if err = container.Bind(&config.AppConfigProvider{}); err != nil {
		panic(err)
	}

	if err := container.Bind(&redis.RedisProvider{}); err != nil {
		panic(err)
	}

	a := container.MustMake(contract.RedisKey).(contract.RedisService)
	client, err := a.GetClient()
	if err != nil {
		panic(err)
	}
	err = client.Set(context.Background(), "demo", "one", 0).Err()
	if err != nil {
		panic(err)
	}

	demoStr, _ := client.Get(context.Background(), "demo").Result()
	fmt.Println("demo: ", demoStr)
}

package main

import (
    "strconv"

    "github.com/PSauerborn/project-gateway/pkg/gateway"
    "github.com/PSauerborn/project-gateway/pkg/utils"
)

var (
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "listen_address": "0.0.0.0",
            "listen_port": "10101",
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100",
            "jwt_secret": "secret_key",
        },
    )
)

func main() {

    port, err := strconv.Atoi(cfg.Get("listen_port"))
    if err != nil {
        panic(err)
    }
    // generate new instance of API gateway with config settings
    router := gateway.NewGateway(cfg.Get("listen_address"), cfg.Get("postgres_url"),
    cfg.Get("jwt_secret"), port, true)
    // defer closing of contexts and run API gateway
    defer router.Shutdown()
    router.Run()
}
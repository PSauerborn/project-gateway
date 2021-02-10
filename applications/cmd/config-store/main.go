package main

import (
    "fmt"
    "strconv"

    "github.com/PSauerborn/project-gateway/pkg/config-store"
    "github.com/PSauerborn/project-gateway/pkg/utils"
)

var (
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "listen_address": "0.0.0.0",
            "listen_port": "10874",
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100",
        },
    )
)

func main() {

    port, err := strconv.Atoi(cfg.Get("listen_port"))
    if err != nil {
        panic(err)
    }
    // generate new instance of API gateway with config settings
    router := config_store.New(cfg.Get("postgres_url"))
    // defer closing of contexts and run API gateway
    conn := fmt.Sprintf("%s:%d", cfg.Get("listen_address"), port)
    router.Run(conn)
}
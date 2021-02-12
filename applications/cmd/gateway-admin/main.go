package main

import (
    "fmt"
    "strconv"

    "github.com/PSauerborn/project-gateway/pkg/gateway-admin"
    "github.com/PSauerborn/project-gateway/pkg/utils"
)

var (
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "listen_address": "0.0.0.0",
            "listen_port": "10101",
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100/project_gateway",
            "jwt_secret": "secret_key",
        },
    )
)

func main() {

    port, err := strconv.Atoi(cfg.Get("listen_port"))
    if err != nil {
        panic(err)
    }
    // generate new instance of admin API with config settings
    router := gateway_admin.NewAdminAPI(cfg.Get("postgres_url"),
        cfg.Get("jwt_secret"), 60)
    router.Run(fmt.Sprintf("%s:%d", cfg.Get("listen_address"), port))
}
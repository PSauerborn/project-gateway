package main

import (
    "fmt"
    "strconv"

    "github.com/PSauerborn/project-gateway/pkg/authentication"
    "github.com/PSauerborn/project-gateway/pkg/utils"
)

var (
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "listen_address": "0.0.0.0",
            "listen_port": "10776",
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100/project_gateway",
            "gateway_admin_redirect_url": "http://0.0.0.0:10101",
        },
    )
)

func main() {

    port, err := strconv.Atoi(cfg.Get("listen_port"))
    if err != nil {
        panic(err)
    }
    // generate new instance of API gateway with config settings
    router := authentication.NewAuthenticationAPI(cfg.Get("postgres_url"),
        cfg.Get("gateway_admin_redirect_url"))
    router.Run(fmt.Sprintf("%s:%d", cfg.Get("listen_address"), port))
}
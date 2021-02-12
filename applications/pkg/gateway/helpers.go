package gateway

import (
    "net/http"
    "net/url"
)

// function used to set proxy headers headers on request
func SetProxyHeaders(request *http.Request, url *url.URL) {
    request.URL.Host = url.Host
    request.URL.Scheme = url.Scheme
    request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
    request.Host = url.Host
}
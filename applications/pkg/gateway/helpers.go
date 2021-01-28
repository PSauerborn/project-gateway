package gateway

import (
    "fmt"
    "net/http"
    "net/url"

    log "github.com/sirupsen/logrus"
    opentracing "github.com/opentracing/opentracing-go"
)

// function used to set proxy headers headers on request
func SetProxyHeaders(request *http.Request, url *url.URL) {
    request.URL.Host = url.Host
    request.URL.Scheme = url.Scheme
    request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
    request.Host = url.Host
}

// function used to set CORS headers on incoming requests
func SetCorsHeaders(response http.ResponseWriter, request *http.Request) {
    origin := request.Header.Get("Origin")
    if len(origin) > 0 {
        response.Header().Set("Access-Control-Allow-Origin", origin)
    } else {
        response.Header().Set("Access-Control-Allow-Origin", "*")
    }
    response.Header().Set("Access-Control-Allow-Credentials", "true")
    response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
    response.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS,GET,PUT,PATCH,DELETE")
}

// function used to extract active jaeger span from request headers
func extractJaegerSpan(req *http.Request) (*opentracing.SpanContext, error) {
    wireContext, err := opentracing.GlobalTracer().Extract(
        opentracing.HTTPHeaders,
        opentracing.HTTPHeadersCarrier(req.Header))
    if err != nil {
        return nil, err
    }
    return &wireContext, nil
}

// function used to set common jaeger tags in span
func setJaegerTags(span opentracing.Span, request *http.Request, user string) {
    log.Debug(fmt.Sprintf("setting tags on jaeger span for user %s", user))
    span.SetTag("http.url", request.URL.Path)
    span.SetTag("http.method", request.Method)
    span.SetTag("uid", user)
}

// function used to create new span if non available or retrieve from headers
func getJaegerSpan(request *http.Request, route string) opentracing.Span {
    var span opentracing.Span
    parentSpan, _ := extractJaegerSpan(request)
    if parentSpan != nil {
        log.Debug("continuing trace with parent jaeger span")
        span = opentracing.StartSpan(route, opentracing.ChildOf(*parentSpan))
    } else {
        log.Debug("starting new jaeger span")
        span = opentracing.StartSpan(route)
    }
    return span
}
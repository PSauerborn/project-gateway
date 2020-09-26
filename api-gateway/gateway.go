package main

import (
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "net/http/httputil"
    log "github.com/sirupsen/logrus"
    opentracing "github.com/opentracing/opentracing-go"
    jaeger "github.com/PSauerborn/jaeger-negroni"
)

func main() {
    // configure environment variables and connect persistence
    ConfigureService()
    ConnectPersistence()

    config := jaeger.Config("jaeger-agent", "api-gateway", 6831)
    tracer := jaeger.NewTracer(config)
    defer tracer.Close()

    log.Info(fmt.Sprintf("starting gateway service at %s:%d", ListenAddress, ListenPort))
    http.HandleFunc("/", Gateway)
    if err := http.ListenAndServe(fmt.Sprintf(":%d", ListenPort), nil); err != nil {
        panic(err)
    }
}

type StandardHTTP struct {}

func(response StandardHTTP) Unauthorized(w http.ResponseWriter) {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func(response StandardHTTP) Forbidden(w http.ResponseWriter) {
    http.Error(w, "Forbidden", http.StatusForbidden)
}

func(response StandardHTTP) InternalServerError(w http.ResponseWriter) {
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func(response StandardHTTP) BadGateway(w http.ResponseWriter) {
    http.Error(w, "Bad Gateway", http.StatusBadGateway)
}

// handler function that acts as API Gateway
func Gateway(response http.ResponseWriter, request *http.Request) {
    // set relevant cors headers
    SetCorsHeaders(response, request)
    // return options calls
    if request.Method == http.MethodOptions {
        return
    }

    log.Info(fmt.Sprintf("received request for URL %s", request.URL.Path))
    // authenticate user using JWToken present in request
    claims, err := AuthenticateUser(request)
    if err != nil {
        log.Error(fmt.Errorf("unable to authenticate user: %v", err))
        StandardHTTP{}.Unauthorized(response)
        return
    }
    log.Info(fmt.Sprintf("received proxy request for user %s", claims.Uid))
    request.Header.Set("X-Authenticated-Userid", claims.Uid)

    // get redirect URL for application from postgres server
    appDetails, err := getProxyURI(request.URL.Path)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve redirect uri: %v", err))
        StandardHTTP{}.BadGateway(response)
        return
    }
    // get jaeger span from current context and defer closure
    span := getJaegerSpan(request, fmt.Sprintf("Proxy - %s", strings.Title(appDetails.ApplicationName)))
    defer span.Finish()
    setJaegerTags(span, request, claims.Uid)

    // inject current span into downstream microservice headers
    opentracing.GlobalTracer().Inject(
        span.Context(),
        opentracing.HTTPHeaders,
        opentracing.HTTPHeadersCarrier(request.Header))

    log.Info(fmt.Sprintf("proxying request to %s", appDetails.RedirectURL))
    // proxy request to relevant microservices
    ProxyRequest(appDetails, response, request)
}

// define function used to proxy request
func ProxyRequest(appDetails ApplicationDetails, response http.ResponseWriter, request *http.Request) {
    // parse URL
    url, _ := url.Parse(appDetails.RedirectURL)
    // set proxy headers on request
    SetProxyHeaders(request, url)
    // create reverse proxy instance and serve request
    proxy := httputil.NewSingleHostReverseProxy(url)
    proxy.ServeHTTP(response, request)
}
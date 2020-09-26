package main

import (
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "net/http/httputil"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "github.com/jackc/pgx/v4"
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

    router := gin.New()
    router.Use(cors.Default())
    router.Use(jaeger.JaegerNegroni(config))

    router.Any("/:application/*proxyPath", Gateway)

    log.Info(fmt.Sprintf("starting gateway service at %s:%d", ListenAddress, ListenPort))
    router.Run(fmt.Sprintf("%s:%d", ListenAddress, ListenPort))
}

// handler function that acts as API Gateway
func Gateway(ctx *gin.Context) {
    // set relevant cors headers
    SetCorsHeaders(ctx.Writer, ctx.Request)
    // return options calls
    if ctx.Request.Method == http.MethodOptions {
        return
    }

    log.Debug(fmt.Sprintf("received request for URL %s", ctx.Request.URL.Path))
    // authenticate user using JWToken present in request
    claims, err := AuthenticateUser(ctx.Request)
    if err != nil {
        log.Error(fmt.Errorf("unable to authenticate user: %v", err))
        StandardHTTP.Unauthorized(ctx)
        return
    }

    // inject uid into X-Authenticated-Userid header
    log.Info(fmt.Sprintf("received proxy request for user %s", claims.Uid))
    ctx.Request.Header.Set("X-Authenticated-Userid", claims.Uid)

    // get redirect URL for application from postgres server
    appDetails, err := persistence.GetModuleDetails(ctx.Param("application"))
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            log.Error(fmt.Sprintf("invalid application %s", ctx.Param("application")))
            StandardHTTP.BadGateway(ctx)
            return
        default:
            log.Error(fmt.Errorf("unable to retrieve redirect uri: %v", err))
            StandardHTTP.InternalServerError(ctx)
            return
        }
    }
    // get jaeger span from current context and defer closure
    span := getJaegerSpan(ctx.Request, fmt.Sprintf("Proxy - %s", strings.Title(appDetails.ApplicationName)))
    defer span.Finish()
    setJaegerTags(span, ctx.Request, claims.Uid)

    // inject current span into downstream microservice headers
    opentracing.GlobalTracer().Inject(
        span.Context(),
        opentracing.HTTPHeaders,
        opentracing.HTTPHeadersCarrier(ctx.Request.Header))

    log.Info(fmt.Sprintf("proxying request to %s", appDetails.RedirectURL))
    // proxy request to relevant microservices
    ProxyRequest(appDetails, ctx.Writer, ctx.Request)
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
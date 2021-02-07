package gateway

import (
    "io"
    "fmt"
    "strings"
    "net/http"
    "net/url"
    "net/http/httputil"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v4"
    log "github.com/sirupsen/logrus"
    opentracing "github.com/opentracing/opentracing-go"
    jaeger "github.com/PSauerborn/jaeger-negroni"
)

type APIGateway struct{
    // configuration settings for API and postgres service
    ListenAddress string
    ListenPort int
    PostgresURL string

    // gin/gonic engine containing routes
    Engine *gin.Engine

    // jaeger tracer used to trace incoming routes
    Tracer io.Closer
}

// function used to create new instance of gateway from connection settings
func NewGateway(listenAddress, postgresUrl, jwtSecret string, listenPort int,
    enableJaegerTracing bool) *APIGateway {

    router := gin.Default()
    // add postgres and authentication middlewares
    router.Use(PostgresSessionMiddleware(postgresUrl))
    router.Use(JWTMiddleware(jwtSecret))

    var tracer io.Closer
    // optionally enable jaeger tracing
    if enableJaegerTracing {
        log.Info("creating new router with jaeger tracing enabled")
        // generate new jaeger config
        config := jaeger.Config("jaeger-agent", "api-gateway", 6831)
        tracer = jaeger.NewTracer(config)
        // add jaeger
        router.Use(jaeger.JaegerNegroni(config))
    }

    // add proxy route to gin router
    router.Any("/:application/*proxyPath", forwardProxy)

    return &APIGateway{
        Engine: router,
        ListenAddress: listenAddress,
        ListenPort: listenPort,
        PostgresURL: postgresUrl,
        Tracer: tracer,
    }
}

// function used to start API router
func(gateway *APIGateway) Run() {
    log.Info(fmt.Sprintf("starting new API Gateway at %s:%d", gateway.ListenAddress,
    gateway.ListenPort))
    // generate connection string and start
    address := fmt.Sprintf("%s:%d",gateway.ListenAddress, gateway.ListenPort)
    gateway.Engine.Run(address)
}

// function used to close any trailing contexts (jaeger, postgres etc)
func(gateway *APIGateway) Shutdown() {
    // close jaeger tracer if tracing is enabled
    if gateway.Tracer != nil {
        gateway.Tracer.Close()
    }
}

// handler function that acts as API Gateway
func forwardProxy(ctx *gin.Context) {

    // inject user ID into downstream headers
    ctx.Request.Header.Set("X-Authenticated-Userid",
        ctx.MustGet("uid").(string))

    // get persistence from middleware context and retrieve app info
    persistence, _ := ctx.MustGet("persistence").(*Persistence)
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
    span := getJaegerSpan(ctx.Request, fmt.Sprintf("Proxy - %s",
    strings.Title(appDetails.ApplicationName)))
    defer span.Finish()
    // set commonly used jaeger tags
    setJaegerTags(span, ctx.Request, ctx.MustGet("uid").(string))

    // inject current span into downstream microservice headers
    opentracing.GlobalTracer().Inject(
        span.Context(),
        opentracing.HTTPHeaders,
        opentracing.HTTPHeadersCarrier(ctx.Request.Header))

    log.Info(fmt.Sprintf("proxying request to %s", appDetails.RedirectURL))
    // proxy request to relevant microservices
    proxyRequest(appDetails, ctx.Writer, ctx.Request)
}

// define function used to proxy request
func proxyRequest(app ApplicationDetails, response http.ResponseWriter, request *http.Request) {
    var redirectUrl string
    // trim app name from URL if specified in app config
    if app.TrimAppName {
        log.Debug(fmt.Sprintf("trimming app name from redirect for application %s", app.ApplicationName))
        replace := fmt.Sprintf("/%s", app.ApplicationName)
        redirectUrl = strings.Replace(app.RedirectURL, replace, "", -1)
    } else {
        redirectUrl = app.RedirectURL
    }

    // construct new URL, set proxy headers and proxy
    redirect, _ := url.Parse(redirectUrl)
    SetProxyHeaders(request, redirect)
    // create reverse proxy instance and serve request
    proxy := httputil.NewSingleHostReverseProxy(redirect)
    proxy.ServeHTTP(response, request)
}
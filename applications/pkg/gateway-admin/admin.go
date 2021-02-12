package gateway_admin

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
    jaeger "github.com/PSauerborn/jaeger-negroni"
)

var (
    cfg *AdminConfig
)

type AdminConfig struct {
    JWTSecret          string
    TokenExpiryMinutes int
}

func NewAdminAPI(postgresUrl, jwtSecret string, tokenExpiry int) *gin.Engine {
    // set global application configuration
    cfg = &AdminConfig{
        JWTSecret: jwtSecret,
        TokenExpiryMinutes: tokenExpiry,
    }

    router := gin.Default()
    router.Use(TimerMiddleware())

    // generate new jaeger config
    config := jaeger.Config("jaeger-agent", "api-gateway-admin", 6831)
    jaeger.NewTracer(config)
    router.Use(jaeger.JaegerNegroni(config))

    // define GET routes
    router.GET("/admin/health_check", healthCheckHandler)

    // add routes to manage gateway (note that admin JWTokens are required)
    router.GET("/admin/applications", PostgresSessionMiddleware(postgresUrl), getApplicationsHandler)
    router.POST("/admin/application", PostgresSessionMiddleware(postgresUrl), addApplicationHandler)

    // define route used to generate gateway token
    router.POST("/admin/token", getTokenHandler)
    return router
}

// function to serve health check route
func healthCheckHandler(ctx *gin.Context) {
    log.Info("received health check request")
    ctx.JSON(http.StatusOK, gin.H{
        "http_code": http.StatusOK, "message": "Service running"})
}

// function to serve current lists of applications
func getApplicationsHandler(ctx *gin.Context) {
    log.Info("received request to retrieve applications")

    db, _ := ctx.MustGet("persistence").(*Persistence)
    apps, err := db.GetAllApplications()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve app list: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "data": apps})
}

// function to add new applications
func addApplicationHandler(ctx *gin.Context) {
    log.Info("received request to create application")
    var request struct {
        AppName     string `json:"app_name" binding:"required"`
        Description string `json:"description" binding:"required"`
        RedirectURL string `json:"redirect_url" binding:"required"`
        TrimAppName *bool  `json:"trim_app_name" binding:"required"`
    }
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("received invalid request body"))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid request body"})
        return
    }

    // extract persistence from context and check if application exists
    db, _ := ctx.MustGet("persistence").(*Persistence)
    exists, err := db.AppExists(request.AppName)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve existing applications: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return

        // if app already exists, return 400
    } else if exists {
        log.Error(fmt.Sprintf("app %s already exists", request.AppName))
        ctx.JSON(http.StatusBadRequest, gin.H{
            "http_code": http.StatusBadRequest, "message": "Application already exists"})
        return
    }

    // create new application in postgres database
    if err := db.CreateNewApplication(request.AppName, request.Description, request.RedirectURL,
        *request.TrimAppName); err != nil {
        log.Error(fmt.Errorf("unable to create new application: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusCreated, gin.H{
        "http_code": http.StatusCreated, "message": "Successfully created application"})
}

// function to generate new JWT token
func getTokenHandler(ctx *gin.Context) {
    log.Info("received request for token")
    var request struct {
        Uid   string `json:"uid"   binding:"required"`
        Admin *bool  `json:"admin" binding:"required"`
    }
    // parse request body from context
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("received invalid request body: %+v", err))
        ctx.JSON(http.StatusBadRequest, gin.H{
            "http_code": http.StatusBadRequest, "message": "Invalid request body"})
        return
    }

    // generate JWToken for user
    token, err := GenerateJWToken(request.Uid, *request.Admin)
    if err != nil {
        log.Error(fmt.Errorf("unable to generate JWToken: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{
        "http_code": http.StatusOK, "token": token})
}
package gateway_admin

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
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
    // define GET routes
    router.GET("/admin/health_check", healthCheckHandler)
    router.GET("/admin/applications", PostgresSessionMiddleware(postgresUrl),
        getApplicationsHandler)
    router.POST("/admin/application", PostgresSessionMiddleware(postgresUrl),
        addApplicationHandler)
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
}

// function to add new applications
func addApplicationHandler(ctx *gin.Context) {
    log.Info("received request to create application")
}

// function to generate new JWT token
func getTokenHandler(ctx *gin.Context) {
    log.Info("received request for token")
    var request struct {
        Uid   string `json:"uid"  binding:"required"`
        Admin *bool `json:"admin" binding:"required"`
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
package config_store

import (
    "fmt"
    "net/http"

    "github.com/google/uuid"
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
)

func New(postgresUrl string) *gin.Engine {
    router := gin.Default()
    // add timer middleware
    router.Use(TimerMiddleware())

    // define GET routes for API
    router.GET("/config-store/health_check", healthCheckHandler)
    router.GET("/config-store/config/:appId", PostgresSessionMiddleware(postgresUrl),
        getAppConfigHandler)
    router.POST("/config-store/config", PostgresSessionMiddleware(postgresUrl),
        addAppConfigHandler)
    router.PATCH("/config-store/config/:appId", PostgresSessionMiddleware(postgresUrl),
        modifyAppConfigHandler)
    router.DELETE("/config-store/config/:appId", PostgresSessionMiddleware(postgresUrl),
        deleteAppConfigHandler)

    return router
}

// function used to serve health check response
func healthCheckHandler(ctx *gin.Context) {
    log.Info("received request to health check route")
    ctx.JSON(http.StatusOK, gin.H{
        "http_code": http.StatusOK, "message": "Service running"})
}

// function to retrieve app config from database
func getAppConfigHandler(ctx *gin.Context) {
    log.Info(fmt.Sprintf("received request for config %s", ctx.Param("appId")))

    // get app ID from path and convert to UUID
    appId, err := uuid.Parse(ctx.Param("appId"))
    if err != nil {
        log.Error(fmt.Errorf("unable to app ID: %+v", err))
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status_code": http.StatusBadRequest, "message": "Invalid app ID"})
        return
    }
    // get config from database and return
    db, _ := ctx.MustGet("db").(*Persistence)
    config, err := db.GetConfigByAppId(appId)
    if err != nil {
        switch err {
        case ErrAppNotFound:
            log.Warn(fmt.Sprintf("cannot find config for app %s", appId))
            ctx.JSON(http.StatusNotFound, gin.H{
                "http_code": http.StatusNotFound, "message": "Cannot find config for app"})
        default:
            log.Error(fmt.Errorf("unable to retrieve config from database: %+v", err))
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "http_code": http.StatusInternalServerError, "message": "Internal server error"})
        }
        return
    }
    ctx.JSON(http.StatusOK, gin.H{
        "http_code": http.StatusOK, "data": config})
}

// function used to add new config to database
func addAppConfigHandler(ctx *gin.Context) {
    log.Info("received request to generate new app config")
    var config struct{
        AppName string                 `json:"app_name" binding:"required"`
        Config  map[string]interface{} `json:"config"   binding:"required"`
    }
    // parse config from request body
    if err := ctx.ShouldBind(&config); err != nil {
        log.Error(fmt.Errorf("unable to parse config from request body: %+v", err))
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status_code": http.StatusBadRequest, "message": "Invalid JSON request body"})
        return
    }

    // insert new config item into database
    db, _ := ctx.MustGet("db").(*Persistence)
    appId, err := db.AddNewConfig(config.AppName, config.Config)
    if err != nil {
        log.Error(fmt.Errorf("unable to register new app config: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "status_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }

    ctx.JSON(http.StatusCreated, gin.H{
        "http_code": http.StatusCreated, "app_id": appId})
}

// function used to modify config from database
func modifyAppConfigHandler(ctx *gin.Context) {
    log.Info(fmt.Sprintf("received request to modify config %s", ctx.Param("appId")))
    var request struct{Operation []map[string]interface{} `json:"operation" binding:"required"`}
     // parse config from request body
     if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("unable to extract JSON Patch from body: %+v", err))
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status_code": http.StatusBadRequest, "message": "Invalid JSON request body"})
        return
    }
    // get app ID from path and convert to UUID
    appId, err := uuid.Parse(ctx.Param("appId"))
    if err != nil {
        log.Error(fmt.Errorf("unable to app ID: %+v", err))
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status_code": http.StatusBadRequest, "message": "Invalid app ID"})
        return
    }

    // insert new config item into database
    db, _ := ctx.MustGet("db").(*Persistence)
    current, err := db.GetConfigByAppId(appId)
    if err != nil {
        switch err {
        case ErrAppNotFound:
            log.Warn(fmt.Sprintf("cannot find config for app %s", appId))
            ctx.JSON(http.StatusNotFound, gin.H{
                "http_code": http.StatusNotFound, "message": "Cannot find config for app"})
        default:
            log.Error(fmt.Errorf("unable to retrieve config from database: %+v", err))
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "http_code": http.StatusInternalServerError, "message": "Internal server error"})
        }
        return
    }

    // perform JSON patch on config
    updated, err := PatchConfig(current, request.Operation)
    if err != nil {
        switch err {
        case ErrInvalidJSONConfig, ErrInvalidPatch:
            log.Warn(fmt.Sprintf("cannot process JSON Patch %+v", err))
            ctx.JSON(http.StatusBadRequest, gin.H{
                "http_code": http.StatusBadRequest, "message": "Invalid JSON Patch Operation"})
        default:
            log.Error(fmt.Errorf("unable to apply JSON Patch: %+v", err))
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "http_code": http.StatusInternalServerError, "message": "Internal server error"})
        }
        return
    }

    // update config in postgres database
    if err := db.UpdateConfigByAppId(appId, updated); err != nil {
        log.Error(fmt.Errorf("unable to updated config in database: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{
        "http_code": http.StatusOK, "message": "Successfully updated config"})
}

// function used to delete config
func deleteAppConfigHandler(ctx *gin.Context) {
    log.Info(fmt.Sprintf("received request to delete config %s", ctx.Param("appId")))

    // get app ID from path and convert to UUID
    appId, err := uuid.Parse(ctx.Param("appId"))
    if err != nil {
        log.Error(fmt.Errorf("unable to app ID: %+v", err))
        ctx.JSON(http.StatusBadRequest, gin.H{
            "status_code": http.StatusBadRequest, "message": "Invalid app ID"})
        return
    }

    db, _ := ctx.MustGet("db").(*Persistence)
    _, err = db.GetConfigByAppId(appId)
    if err != nil {
        switch err {
        case ErrAppNotFound:
            log.Warn(fmt.Sprintf("cannot find config for app %s", appId))
            ctx.JSON(http.StatusNotFound, gin.H{
                "http_code": http.StatusNotFound, "message": "Cannot find config for app"})
        default:
            log.Error(fmt.Errorf("unable to retrieve config from database: %+v", err))
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "http_code": http.StatusInternalServerError, "message": "Internal server error"})
        }
        return
    }

    if err := db.DeleteConfigByAppId(appId); err != nil {
        log.Error(fmt.Errorf("unable to delete config: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{
        "http_code": http.StatusOK, "message": "Successfully delete config"})
}
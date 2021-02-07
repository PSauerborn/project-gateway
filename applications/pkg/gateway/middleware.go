package gateway

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
)

// middleware used to inject postgres connection into request
func PostgresSessionMiddleware(postgresUrl string) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // create new persistence instance and connect to postgres
        db := NewPersistence(postgresUrl)
        conn, err := db.Connect()
        if err != nil {
            log.Error(fmt.Errorf("unable to retrieve assets from postgres: %+v", err))
            StandardHTTP.InternalServerError(ctx)
            return
        }
        defer conn.Close()

        ctx.Set("persistence", db)
        ctx.Next()
    }
}

// middleware used to parse JWTokens from request
func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // set relevant cors headers and return options calls
        SetCorsHeaders(ctx.Writer, ctx.Request)
        if ctx.Request.Method == http.MethodOptions {
            log.Debug("received options calls. returning...")
            ctx.AbortWithStatus(http.StatusOK)
            return
        }

        log.Debug(fmt.Sprintf("received request for URL %s", ctx.Request.URL.Path))
        // authenticate user using JWToken present in request
        claims, err := authenticateUser(ctx.Request, jwtSecret)
        if err != nil {
            log.Error(fmt.Errorf("unable to authenticate user: %v", err))
            StandardHTTP.Unauthorized(ctx)
            return
        }

        // inject uid into request context
        log.Info(fmt.Sprintf("received proxy request for user %s", claims.Uid))
        ctx.Set("uid", claims.Uid)

        ctx.Next()
    }
}
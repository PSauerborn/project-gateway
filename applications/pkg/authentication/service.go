package authentication

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"

    jaeger "github.com/PSauerborn/jaeger-negroni"
)

var cfg *AuthConfig

type AuthConfig struct {
    TokenRedirectURL string
}

// function to create new authentication service
func NewAuthenticationAPI(postgresUrl, tokenRedirectUrl string) *gin.Engine {
    // set global configuration for module
    cfg = &AuthConfig{
        TokenRedirectURL: tokenRedirectUrl,
    }
    // generate new instance of gin router and add routes
    router := gin.Default()
    router.Use(TimerMiddleware())

    // generate new jaeger config
    config := jaeger.Config("jaeger-agent", "identity-provider", 6831)
    jaeger.NewTracer(config)
    router.Use(jaeger.JaegerNegroni(config))

    router.GET("/health_check", healthCheckHandler)
    // define POST routes
    router.POST("/signup", PostgresSessionMiddleware(postgresUrl),
        signUpHandler)
    router.POST("/token", PostgresSessionMiddleware(postgresUrl),
        loginHandler)

    return router
}

func healthCheckHandler(ctx *gin.Context) {
    log.Info("received request to health check route")
    ctx.JSON(http.StatusOK, gin.H{
        "http_code": http.StatusOK, "message": "Service running"})
}

// function to create new user in database
func signUpHandler(ctx *gin.Context) {
    log.Info("received signup request")
    var request struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
        Email    string `json:"email"    binding:"required"`
    }
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("received invalid request body: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid request body"})
        return
    }

    db, _ := ctx.MustGet("persistence").(*Persistence)
    exists, err := db.UserExists(request.Username)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve list of users: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }

    // check email validity with regex
    if !isValidEmail(request.Email) {
        log.Error(fmt.Sprintf("received invalid email address %s", request.Email))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid email address"})
        return
    }

    // return 400 response if user already exists
    if exists {
        log.Error(fmt.Errorf("user %s already exists", request.Username))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "User already exists"})
        return
    }
    // insert new user into database tables
    if err := db.CreateUser(request.Username, request.Password,
        request.Email); err != nil {
        log.Error(fmt.Errorf("unable to create new user: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusCreated,
        gin.H{"http_code": http.StatusCreated, "message": "Successfully created user"})
}

func loginHandler(ctx *gin.Context) {
    log.Info("received login request")
    var request struct {
        Uid      string `json:"uid"      binding:"required"`
        Password string `json:"password" binding:"required"`
    }
    // parse request body and return any errors
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("received invalid request body: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid request body"})
        return
    }

    db, _ := ctx.MustGet("persistence").(*Persistence)
    // check if user is authenticated
    creds, err := authenticateUser(db, request.Uid, request.Password)
    if err != nil {
        switch err {
        case ErrUserDoesNotExist, ErrInvalidPassword:
            log.Warn(fmt.Sprintf("received unauthenticated request from user %s", request.Uid))
            ctx.JSON(http.StatusUnauthorized,
                gin.H{"http_code": http.StatusUnauthorized, "message": "Unauthorized"})
        default:
            log.Error(fmt.Sprintf("unable to authenticate user: %+v", err))
            ctx.JSON(http.StatusInternalServerError,
                gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        }
        return
    }

    // get access token from gateway
    token, err := getAccessToken(creds)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve access token for user %s: %+v", request.Uid, err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"http_code": http.StatusOK, "token": token})
}
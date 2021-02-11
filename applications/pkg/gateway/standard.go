package gateway

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

var (
    StandardHTTP = StandardJSONResponse{}
)

// define interface used to store a collection of standard HTTP responses
type StandardHTTPResponse interface{
    Success(ctx *gin.Context)
    InvalidRequestBody(ctx *gin.Context)
    InvalidJSON(ctx *gin.Context)
    InvalidRequest(ctx *gin.Context)
    NotFound(ctx *gin.Context)
    Unauthorized(ctx *gin.Context)
    Forbidden(ctx *gin.Context)
    InternalServerError(ctx *gin.Context)
    BadGateway(ctx *gin.Context)
}

// define set of standard HTTP Responses in JSON format
type StandardJSONResponse struct{}

func(response StandardJSONResponse) Success(ctx *gin.Context) {
    ctx.JSON(http.StatusOK, gin.H{
        "http_code": http.StatusOK,
        "success": true,
        "message": "Success" })
}

func(response StandardJSONResponse) InvalidRequestBody(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
        "http_code": http.StatusBadRequest,
        "success": false,
        "message": "Invalid request body" })
}

func(response StandardJSONResponse) InvalidJSON(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
        "http_code": http.StatusBadRequest,
        "success": false,
        "message": "Request body must be JSON serializable" })
}

func(response StandardJSONResponse) InvalidRequest(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
        "http_code": http.StatusBadRequest,
        "success": false,
        "message":
        "Invalid request" })
}

func(response StandardJSONResponse) Unauthorized(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
        "http_code": http.StatusUnauthorized,
        "success": false,
        "message": "Unauthorized" })
}

func(response StandardJSONResponse) Forbidden(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
        "http_code": http.StatusForbidden,
        "success": false,
        "message": "Access forbidden" })
}

func(response StandardJSONResponse) NotFound(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
        "http_code": http.StatusNotFound,
        "success": false,
        "message": "Not found" })
}

func(response StandardJSONResponse) InternalServerError(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
        "http_code": http.StatusInternalServerError,
        "success": false,
        "message": "Internal server error" })
}

func(response StandardJSONResponse) BadGateway(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
        "http_code": http.StatusBadGateway,
        "success": false, "message":
        "Bad gateway" })
}

func(response StandardJSONResponse) FeatureNotSupported(ctx *gin.Context) {
    ctx.AbortWithStatusJSON(http.StatusNotImplemented, gin.H{
        "http_code": http.StatusNotImplemented,
        "success": false,
        "message": "Feature not implemented" })
}


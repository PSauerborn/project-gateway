package utils

import (
    "net/http"
)

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
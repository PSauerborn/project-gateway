package gateway

import (
    "fmt"
    "strings"
    "errors"
    "net/http"

    "github.com/dgrijalva/jwt-go"
    log "github.com/sirupsen/logrus"
)

var (
    // define custom errors returned by application
    ErrInvalidJWToken     = errors.New("Invalid JWToken")
    ErrInvalidTokenSchema = errors.New("Invalid Token Schema")
    ErrInvalidAuthHeader  = errors.New("Invalid authorization header")
)

// function used to parse JWT token string into JWT claims
func parseJWToken(tokenString, secret string) (*JWTClaims, error) {
    log.Debug(fmt.Sprintf("parsing JWToken %s", tokenString))
    // parse token using JWT secret
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })
    // parse token into custom claims object
    if customClaims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
        return customClaims, nil
    } else {
        log.Error(fmt.Errorf("unable to parse JWT claims: %v", err))
        return nil, ErrInvalidJWToken
    }
}

// function used to extract bearer token from header. note that bearer
// token must be in the Authorization: Bearer <token> format, else
// an error is returned
func extractToken(header string) (string, error) {
    if strings.HasPrefix(header, "Bearer ") {
        return header[7:], nil
    }
    return "", ErrInvalidTokenSchema
}

// function used to authenticate incoming users. access tokens are
// pulled from the Authorization: Bearer <token> header, and then
// parsed using JWT secret define in application. claims are returned
// along with any possible errors (header format, missing header etc)
func authenticateUser(request *http.Request, secret string) (*JWTClaims, error) {
    // extract token from authentication header
    tokenString, err := extractToken(request.Header.Get("Authorization"))
    if err != nil || tokenString == "undefined" {
        log.Error(fmt.Errorf("invalid authorization header: %v", err))
        return nil, ErrInvalidAuthHeader
    }

    // parse token claims using token string and secret
    tokenClaims, err := parseJWToken(tokenString, secret)
    if err != nil {
        log.Error(fmt.Errorf("unable to parse JWToken: %v", err))
        return nil, err
    }
    return tokenClaims, nil
}

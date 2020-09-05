package main

import (
	"fmt"
	"strings"
	"errors"
	"net/http"
	"net/url"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)


type JWTClaims struct {
	Uid   string      `json:"uid"`
	Admin bool	      `json:"admin"`
	jwt.StandardClaims
}

// function used to parse JWT token
func ParseJWToken(tokenString string) (*JWTClaims, error) {
	log.Info(fmt.Sprintf("parsing JWToken %s", tokenString))
	// parse token using JWT secret
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})
	// parse token into custom claims object
	if customClaims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return customClaims, nil
	} else {
		log.Error(fmt.Errorf("unable to parse JWT claims: %v", err))
		return nil, errors.New("invalid JWToken")
	}
}

// define function used to extract token from request header
func ExtractToken(header string) (string, error) {
	if strings.HasPrefix(header, "Bearer ") {
		return header[7:], nil
	}
	return "", errors.New(fmt.Sprintf("received invalid token schema: %s", header))
}

// define function used to authenticate user
func AuthenticateUser(request *http.Request) (*JWTClaims, error) {
	// extract token from authentication header
	tokenString, err := ExtractToken(request.Header.Get("Authorization"))
	if err != nil || tokenString == "undefined" {
		log.Error(fmt.Errorf("invalid authorization header: %v", err))
		return nil, errors.New("received invalid authorization header")
	}
	// parse token claims using token string
	tokenClaims, err := ParseJWToken(tokenString)
	if err != nil {
		log.Error(fmt.Errorf("unable to parse JWToken: %v", err))
		return nil, err
	}
	return tokenClaims, nil
}

// define function used to determine relevant proxy URL
func getProxyURI(path string) (ApplicationDetails, error) {
	// pass request path and extract application name
	params := strings.Split(path, "/")
	if len(params) < 2 {
		log.Error("received request to empty route")
		return ApplicationDetails{}, errors.New("invalid request path")
	}
	log.Info(fmt.Sprintf("getting redirect for application '%s'", params[1]))

	// get application details from postgres server
	details, err := GetModuleDetails(params[1])
	if err != nil {
		return ApplicationDetails{}, err
	}
	return details, nil
}

// function used to set proxy headers headers on request
func SetProxyHeaders(request *http.Request, url *url.URL) {
	request.URL.Host = url.Host
	request.URL.Scheme = url.Scheme
	request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
	request.Host = url.Host
}

func SetCorsHeaders(response http.ResponseWriter, request *http.Request) {
	origin := request.Header.Get("Origin")
	if len(origin) > 0 {
		response.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		response.Header().Set("Access-Control-Allow-Origin", "*")
	}
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.Header().Set("Access-Control-Allow-Credentials", "true")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	response.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS,GET,PUT,PATCH,DELETE")
}
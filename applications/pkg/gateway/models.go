package gateway

import (
    "time"

    "github.com/google/uuid"
    "github.com/dgrijalva/jwt-go"
)

type ApplicationDetails struct {
    ApplicationID   uuid.UUID `json:"application_id"`
    ApplicationName string 	  `json:"application_name"`
    CreatedAt		time.Time `json:"created_at"`
    Description	    string    `json:"description"`
    RedirectURL 	string 	  `json:"redirect_url"`
    TrimAppName     bool	  `json:"trim_app_name"`
}

type JWTClaims struct {
    Uid   string      `json:"uid"`
    Admin bool	      `json:"admin"`
    jwt.StandardClaims
}

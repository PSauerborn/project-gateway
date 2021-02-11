package gateway_admin

import (
    "time"

    "github.com/dgrijalva/jwt-go"
)

// function used to generate JWToken with UID and expiry date
func GenerateJWToken(uid string, admin bool) (string, error) {
    // evaluate expiry time
    expiry := time.Now().UTC()
    expiry = expiry.Add(time.Duration(cfg.TokenExpiryMinutes) * time.Minute)
    // generate token and sign with secret key
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "uid": uid,
        "exp": expiry.Unix(),
        "admin": admin,
    })
    return token.SignedString([]byte(cfg.JWTSecret))
}

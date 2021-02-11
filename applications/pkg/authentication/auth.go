package authentication

import (
    "fmt"
    "net"
    "regexp"
    "strings"
    "errors"
    "bytes"
    "net/http"
    "encoding/json"

    "golang.org/x/crypto/bcrypt"
    log "github.com/sirupsen/logrus"
)

var (
    // define custom errors
    ErrInvalidPassword        = errors.New("Unable to authenticate user: invalid password")
    ErrUserDoesNotExist       = errors.New("Unable to authenticate user: user does not exist")
    ErrInvalidGatewayResponse = errors.New("Received invalid response from API gateway")

    // define regex to validate emails
    emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// function used to authenticate user by comparing password
// against password hash stored in database
func authenticateUser(db *Persistence, username, password string) (UserDetails, error) {
    log.Debug(fmt.Sprintf("authenticting user %s", username))
    // get user details from postgres database
    user, err := db.GetUserDetails(username)
    if err != nil {
        return UserDetails{}, err
    }
    // compare password to hashed password stored in database
    if !comparePasswords(password, user.HashedPassword) {
        return user, ErrInvalidPassword
    }
    return user, nil
}

// function used to hash and salt user passwords
func hashAndSalt(password string) string {
    // convert passwords into byte array and hash
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Error(err)
    }
    return string(hash)
}

// function used to compare password to hashed password from database
func comparePasswords(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err != nil {
        log.Warn(err)
        return false
    }
    return true
}

// function used to get access token from API gateway
func getAccessToken(user UserDetails) (string, error) {
    log.Debug("retrieveing access token from API gateway")

    // generate JSON payload for POST request
    body, _ := json.Marshal(map[string]interface{}{
        "uid": user.Username,
        "admin": user.Admin})

    // createnew HTTP instance and set request headers
    req, err := http.NewRequest("POST", fmt.Sprintf("%s/admin/token", cfg.TokenRedirectURL),
        bytes.NewBuffer(body))
    if err != nil {
        log.Error(fmt.Errorf("unable to generate new HTTP Request: %+v", err))
        return "", err
    }
    // set JSON as content type
    req.Header.Set("Content-Type", "application/json")

    // generate new HTTP client and execute request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute HTTP request: %+v", err))
        return "", err
    }
    defer resp.Body.Close()

    // handle response based on status code
    switch resp.StatusCode {
    case 200:
        log.Debug("successfully retrieved business data")
        var response struct{
            HTTPCode int    `json:"http_code"`
            Token 	 string `json:"token"`
        }
        // parse JSON response from request
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            log.Error(fmt.Errorf("received invalid response from API Gateway: %+v", err))
            return "", ErrInvalidGatewayResponse
        }
        return response.Token, nil
    default:
        log.Error(fmt.Errorf("received invalid response from API gateway %d", resp.StatusCode))
        return "", ErrInvalidGatewayResponse
    }
}

// function used to check if email address
func isValidEmail(email string) bool {
    if len(email) < 3 && len(email) > 254 {
        return false
    }
    // check that regex matches
    if !emailRegex.MatchString(email) {
        return false
    }
    parts := strings.Split(email, "@")
    // execute MX record lookup to validate
    mx, err := net.LookupMX(parts[1])
    if err != nil || len(mx) == 0 {
        return false
    }
    return true
}
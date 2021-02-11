package authentication

import (
    "fmt"
    "time"
    "context"

    "github.com/jackc/pgx/v4"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"

    "github.com/PSauerborn/project-gateway/pkg/utils"
)

type Persistence struct {
    *utils.Persistence
}

// function to generate new persistence layer
func NewPersistence(postgresUrl string) *Persistence {
    // create instance of base persistence
    basePersistence := utils.NewPersistence(postgresUrl)
    return &Persistence{
        basePersistence,
    }
}

type UserDetails struct {
    Uid            uuid.UUID
    HashedPassword string
    Username       string
    Email          string
    Created        time.Time
    Admin          bool
}

// function to retrieve user details from database
func(db *Persistence) GetUserDetails(username string) (UserDetails, error) {
    log.Debug(fmt.Sprintf("fetching user details for user %+s", username))

    query := `SELECT users.uid,users.password,details.email,details.created,
        details.admin FROM users users INNER JOIN user_details details ON users.uid
        = details.uid WHERE users.username = $1`
    row := db.Session.QueryRow(context.Background(), query, username)

    var (hashedPassword, email string; created time.Time; admin bool; userId uuid.UUID)
    // read values from cursor into local variables
    if err := row.Scan(&userId, &hashedPassword, &email, &created, &admin); err != nil {
        switch err {
        case pgx.ErrNoRows:
            return UserDetails{}, ErrUserDoesNotExist
        default:
            log.Error(fmt.Errorf("unable to read data into local variables: %+v", err))
            return UserDetails{}, err
        }
    }

    details := UserDetails{
        Uid: userId,
        HashedPassword: hashedPassword,
        Username: username,
        Email: email,
        Created: created,
        Admin: admin,
    }
    return details, nil
}

// function used to determine is a user already exists in the database
func(db *Persistence) UserExists(username string) (bool, error) {
    log.Debug(fmt.Sprintf("checking if user %s exists...", username))
    _, err := db.GetUserDetails(username)
    if err != nil {
        switch err {
        case ErrUserDoesNotExist:
            return false, nil
        default:
            return false, err
        }
    }
    return true, nil
}

func(db *Persistence) CreateUser(username, password, email string) error {
    log.Debug(fmt.Sprintf("inserting new user entry foruser %s...", username))
    uid := uuid.New()

    var query string
    query = `INSERT INTO users(uid,username,password) VALUES($1,$2,$3)`
    _, err := db.Session.Exec(context.Background(), query, uid, username, hashAndSalt(password))
    if err != nil {
        log.Error(fmt.Sprintf("unable to insert user into users table: %+v", err))
        return err
    }

    query = `INSERT INTO user_details(uid,email) VALUES($1,$2)`
    _, err = db.Session.Exec(context.Background(), query, uid, email)
    if err != nil {
        log.Error(fmt.Sprintf("unable to insert user into details table: %+v", err))
        return err
    }
    return nil
}
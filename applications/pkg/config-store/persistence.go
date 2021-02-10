package config_store

import (
    "fmt"
    "context"
    "errors"
    "encoding/json"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v4"
    "github.com/jackc/pgx/v4/pgxpool"
    log "github.com/sirupsen/logrus"
)


var (
    // define custom errors
    ErrAppNotFound       = errors.New("Cannot find config for given application")
    ErrInvalidJSONConfig = errors.New("Cannot parse config to JSON")
)

type Persistence struct{
    DatabaseURL string
    Session     *pgxpool.Pool
}

// function to connect persistence to postgres server
// note that the connection is returned and should be
// closed with a defer conn.Close(context) statement
func(db *Persistence) Connect() (*pgxpool.Pool, error) {
    log.Debug(fmt.Sprintf("creating new database connection"))
    // connect to postgres server and set session in persistence
    conn, err := pgxpool.Connect(context.Background(), db.DatabaseURL)
    if err != nil {
        log.Error(fmt.Errorf("error connecting to postgres service: %+v", err))
        return nil, err
    }
    db.Session = conn
    return conn, err
}

func NewPersistence(url string) *Persistence {
    return &Persistence{
        DatabaseURL: url,
    }
}

// function to add new config into postgres database
func(db *Persistence) AddNewConfig(appName string, config map[string]interface{}) (uuid.UUID, error) {
    log.Debug("adding new config to database...")

    var err error
    appId := uuid.New()
    // convert config to JSON string
    configString, err := json.Marshal(config)
    if err != nil {
        log.Error(fmt.Errorf("unable to convert config to JSON string: %+v", err))
        return appId, err
    }

    // insert new config into database
    query := `INSERT INTO app_configs(app_id,app_name,config) VALUES($1,$2,$3)`
    _, err = db.Session.Exec(context.Background(), query, appId, appName, configString)
    if err != nil {
        log.Error(fmt.Errorf("unable to insert config into database: %+v", err))
        return appId, err
    }
    return appId, nil
}

// function used to get config from database based on App ID
func(db *Persistence) GetConfigByAppId(appId uuid.UUID) (map[string]interface{}, error) {
    log.Debug(fmt.Sprintf("fetching config for app %s...", appId))

    query := `SELECT app_name, config FROM app_configs WHERE app_id=$1`
    row := db.Session.QueryRow(context.Background(), query, appId)

    var (appName string; config map[string]interface{})
    // scan data into database
    if err := row.Scan(&appName, &config); err != nil {
        switch err {
        case pgx.ErrNoRows:
            return nil, ErrAppNotFound
        default:
            log.Error(fmt.Errorf("unable to scan data into local variables: %+v", err))
            return nil, err
        }
    }
    return config, nil
}

// function used update a config in the database based on app ID
func(db *Persistence) UpdateConfigByAppId(appId uuid.UUID, updated map[string]interface{}) error {
    log.Debug(fmt.Sprintf("updating config for app %s...", appId))

    // convert config string to JSON before returning
    configString, err := json.Marshal(updated)
    if err != nil {
        log.Error(fmt.Errorf("unable to parse config to JSON format: %+v", err))
        return ErrInvalidJSONConfig
    }

    query := `UPDATE app_configs SET config=$1 WHERE app_id=$2`
    _, err = db.Session.Exec(context.Background(), query, configString, appId)
    return err
}

// function used delete a config in the database based on app ID
func(db *Persistence) DeleteConfigByAppId(appId uuid.UUID) error {
    log.Debug(fmt.Sprintf("deleting config for app %s...", appId))

    query := `DELETE FROM app_configs WHERE app_id=$1`
    _, err := db.Session.Exec(context.Background(), query, appId)
    return err
}
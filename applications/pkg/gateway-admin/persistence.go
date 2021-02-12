package gateway_admin

import (
    "fmt"
    "time"
    "context"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v4"
    log "github.com/sirupsen/logrus"

    "github.com/PSauerborn/project-gateway/pkg/utils"
)

type Persistence struct {
    *utils.Persistence
}

func NewPersistence(postgresUrl string) *Persistence {
    // create instance of base persistence
    basePersistence := utils.NewPersistence(postgresUrl)
    return &Persistence{
        basePersistence,
    }
}

type ApplicationDetails struct {
    ApplicationId   uuid.UUID `json:"application_id"`
    ApplicationName string    `json:"application_name"`
    Description     string    `json:"description"`
    CreatedAt       time.Time `json:"created_at"`
    RedirectURL     string    `json:"redirect_url"`
    TrimAppName     bool      `json:"trim_app_name"`
}

func(db *Persistence) GetAllApplications() ([]ApplicationDetails, error) {
    log.Debug("fetching applications from database...")

    apps := []ApplicationDetails{}
    query := `SELECT application_id,application_name,description,
        created_at,redirect_url,trim_app_name FROM applications`
    rows, err := db.Session.Query(context.Background(), query)
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            return apps, nil
        default:
            return apps, err
        }
    }

    for rows.Next() {
        var (appName, description, redirect string; trimName bool; created time.Time; appId uuid.UUID)
        if err := rows.Scan(&appId, &appName, &description, &created,
            &redirect, &trimName); err != nil {
            log.Warn(fmt.Errorf("unable to scan data into local variables: %+v", err))
            continue
        }
        apps = append(apps, ApplicationDetails{
            ApplicationId: appId,
            ApplicationName: appName,
            Description: description,
            CreatedAt: created,
            RedirectURL: redirect,
            TrimAppName: trimName,
        })
    }
    return apps, nil
}

// function to check if application exists
func(db *Persistence) AppExists(appName string) (bool, error) {
    log.Debug(fmt.Sprintf("checking if application %s exists...", appName))
    query := `SELECT application_id FROM applications WHERE application_name = $1`
    row := db.Session.QueryRow(context.Background(), query, appName)

    var appId uuid.UUID
    if err := row.Scan(&appId); err != nil {
        switch err {
        case pgx.ErrNoRows:
            return false, nil
        default:
            return false, err
        }
    }
    return true, nil
}

// function to insert new application into database
func(db *Persistence) CreateNewApplication(appName, description, redirect string,
    trimName bool) error {

    log.Debug("adding new application to datbase...")
    appId := uuid.New()

    query := `INSERT INTO applications(application_id,application_name,description,
        redirect_url,trim_app_name) VALUES($1,$2,$3,$4,$5)`
    _, err := db.Session.Exec(context.Background(), query, appId, appName, description,
        redirect, trimName)
    return err
}
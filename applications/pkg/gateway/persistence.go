package gateway

import (
    "fmt"
    "time"
    "context"

    "github.com/jackc/pgx/v4/pgxpool"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"
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

// function used to retrieve module details from database
func(db *Persistence) GetModuleDetails(application string) (ApplicationDetails, error) {
    log.Debug(fmt.Sprintf("fetching module details for module %s", application))

    var (applicationId uuid.UUID; description, redirectUrl string; created time.Time; trim bool)

    query := `SELECT application_id,created_at,description,redirect_url,trim_app_name
        FROM applications WHERE application_name=$1`
    // get module details from postgres server
    results := db.Session.QueryRow(context.Background(), query, application)
    err := results.Scan(&applicationId, &created, &description, &redirectUrl, &trim)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve module details: %v", err))
        return ApplicationDetails{}, err
    }
    details := ApplicationDetails{
        ApplicationID: applicationId,
        ApplicationName: application,
        CreatedAt: created,
        Description: description,
        RedirectURL: redirectUrl,
        TrimAppName: trim,
    }
    return details, nil
}
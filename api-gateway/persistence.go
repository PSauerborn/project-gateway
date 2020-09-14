package main

import (
    "fmt"
    "time"
    "context"
    "github.com/jackc/pgx/v4/pgxpool"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"
)

var persistence *Persistence

type Persistence struct {
    conn *pgxpool.Pool
}

// function used to connect postgres connection
func ConnectPersistence() {
    log.Info(fmt.Sprintf("attempting postgres connection with connection string %s", PostgresConnection))
    db, err := pgxpool.Connect(context.Background(), PostgresConnection)
    if err != nil {
        log.Fatal(fmt.Errorf("unable to connect to postgres server: %v", err))
    }
    log.Info("successfully connected to postgres")
    // connect persistence and assign to persistence var
    persistence = &Persistence{ db }
}

func(db Persistence) GetModuleDetails(application string) (ApplicationDetails, error) {

    log.Debug(fmt.Sprintf("fetching module details for module %s", application))
    var (applicationId uuid.UUID; description, redirectUrl string; created time.Time; trim bool)

    // get module details from postgres server
    results := db.conn.QueryRow(context.Background(), "SELECT application_id,created_at,description,redirect_url,trim_app_name FROM applications WHERE application_name=$1", application)
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
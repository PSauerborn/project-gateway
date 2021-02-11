package gateway

import (
    "fmt"
    "time"
    "context"

    "github.com/google/uuid"
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
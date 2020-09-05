package main

import (
	"fmt"
	"time"
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var (
	PostgresConnection = OverrideStringVariable("POSTGRES_CONNECTION", "postgres://postgres:postgres-dev@localhost:5432/gateway")
)

type ApplicationDetails struct {
	ApplicationID   uuid.UUID `json:"application_id"`
	ApplicationName string 	  `json:"application_name"`
	CreatedAt		time.Time `json:"created_at"`
	Description	    string    `json:"description"`
	RedirectURL 	string 	  `json:"redirect_url"`
	TrimAppName     bool	  `json:"trim_app_name"`
}

func GetModuleDetails(module string) (ApplicationDetails, error) {
	db, err := pgx.Connect(context.Background(), PostgresConnection)
	if err != nil {
		log.Error(fmt.Errorf("unable to connect to Postgres Server: %v", err))
		return ApplicationDetails{}, err
	}
	defer db.Close(context.Background())

	var (applicationId uuid.UUID; applicationName, description, redirectUrl string; created time.Time; trim bool)

	// get module details from postgres server
	results := db.QueryRow(context.Background(), "SELECT application_id,application_name,created_at,description,redirect_url,trim_app_name FROM applications WHERE application_name=$1", module)
	err = results.Scan(&applicationId, &applicationName, &created, &description, &redirectUrl, &trim)
	if err != nil {
		log.Error(fmt.Errorf("unable to retrieve module details: %v", err))
		return ApplicationDetails{}, err
	}

	details := ApplicationDetails{
		ApplicationID: applicationId,
		ApplicationName: applicationName,
		CreatedAt: created,
		Description: description,
		RedirectURL: redirectUrl,
		TrimAppName: trim,
	}
	return details, nil
}
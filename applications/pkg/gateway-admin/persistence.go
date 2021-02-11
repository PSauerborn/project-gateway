package gateway_admin

import (
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

func(db *Persistence) GetAllApplications() error {
    log.Debug("fetching applications from database...")
    return nil
}
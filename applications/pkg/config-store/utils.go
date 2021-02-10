package config_store

import (
    "fmt"
    "errors"
    "encoding/json"

    log "github.com/sirupsen/logrus"
    jsonpatch "github.com/evanphx/json-patch"
)

var (
    // define custom errors
    ErrInvalidPatch      = errors.New("Invalid JSON patch operation")
    ErrInvalidConfigJSON = errors.New("Invalid config JSON")
)

// function used to perform JSON patch operation on a
// config map for a given application
func PatchConfig(config map[string]interface{},
    operation []map[string]interface{}) (map[string]interface{}, error) {

    // convert operation to JSON format and parse into JSON patch operation
    patchJson, err := json.Marshal(operation)
    if err != nil {
        log.Error(fmt.Errorf("unable to convert patch operation to JSON: %+v", err))
        return map[string]interface{}{}, ErrInvalidPatch
    }
    patch, err := jsonpatch.DecodePatch(patchJson)
    if err != nil {
        log.Error(fmt.Errorf("unable to parse Json Patch operation: %+v", err))
        return map[string]interface{}{}, ErrInvalidPatch
    }

    // convert retrieved config file into JSON string if not null
    var configJson []byte
    if config == nil {
        configJson = []byte(`{}`)
    } else {
        configJson, err = json.Marshal(config)
        if err != nil {
            log.Error(fmt.Errorf("unable to convert config to JSON: %+v", err))
            return map[string]interface{}{}, ErrInvalidConfigJSON
        }
    }

    // apply JSON patch operation to stringified config instance
    modified, err := patch.Apply(configJson)
    if err != nil {
        log.Error(fmt.Errorf("unable to apply JSON patch: %+v", err))
        return map[string]interface{}{}, ErrInvalidPatch
    }

    log.Debug(fmt.Sprintf("successfully applied JSON patch to config: %s", modified))
    // convert final JSON string back to interface before returning
    var cfg map[string]interface{}
    if err := json.Unmarshal(modified, &cfg); err != nil {
        return cfg, ErrInvalidConfigJSON
    }
    return cfg, nil
}
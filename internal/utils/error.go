package utils

import (
	"encoding/json"
	"fmt"
	"log"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func LogError(err error, msg string) {
	if err != nil {
		logger := GetLogger("utils")
		logger.ErrorWithErr(msg, err)
	}
}

func ParseErrToType(err error, target interface{}) error {
	errString := err.Error()
	if errString == "" {
		return nil
	}

	er := json.Unmarshal([]byte(errString), target)
	if er != nil {
		return fmt.Errorf("failed to parse JSON: %w", er)
	}

	return nil
}
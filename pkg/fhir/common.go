package fhir

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func check(err error) {
	if err == nil {
		return
	}

	log.Error(err)
}

func checkFatal(err error) {
	if err == nil {
		return
	}

	log.Error(err)
	os.Exit(1)
}

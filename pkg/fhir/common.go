package fhir

import (
	log "github.com/sirupsen/logrus"
)

func check(err error) {
	if err == nil {
		return
	}

	log.Error(err)
}

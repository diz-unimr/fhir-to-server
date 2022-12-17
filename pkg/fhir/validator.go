package fhir

import (
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	log "github.com/sirupsen/logrus"
	"os"
)

func ParseBundle(fhirData []byte) {
	bundle, err := fhir.UnmarshalBundle(fhirData)
	check(err)

	log.Info(bundle)
}

func check(err error) {
	if err == nil {
		return
	}

	log.WithError(err).Error("Terminating")
	os.Exit(1)
}

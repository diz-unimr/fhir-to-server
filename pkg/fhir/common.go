package fhir

import (
	"github.com/rs/zerolog/log"
)

func check(err error) {
	if err == nil {
		return
	}

	log.Error().Err(err)
}

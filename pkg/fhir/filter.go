package fhir

import (
	"encoding/json"
	"fhir-to-server/pkg/config"
	models "github.com/samply/golang-fhir-models/fhir-models/fhir"
	"time"
)

type DateFilter struct {
	Date       time.Time
	Comparator func(t time.Time) bool
	Inclusive  bool
}

func NewDateFilter(config config.DateConfig) *DateFilter {
	loc, _ := time.LoadLocation("Europe/Berlin")
	date, err := time.ParseInLocation("2006-01-02", config.Value, loc)
	checkFatal(err)

	var comp func(t time.Time) bool
	if config.Comparator == ">" {
		comp = date.Before
	} else if config.Comparator == "<" {
		comp = date.After
	}

	return &DateFilter{Date: date, Comparator: comp, Inclusive: config.Inclusive}
}

type ResourceTypeDto struct {
	Type *string `bson:"resourceType" json:"resourceType"`
}

type Period struct {
	Start *string `bson:"start,omitempty" json:"start,omitempty"`
	End   *string `bson:"end,omitempty" json:"end,omitempty"`
}

type DateTimeResource struct {
	Type              *string `bson:"resourceType" json:"resourceType"`
	EffectiveDateTime *string `bson:"effectiveDateTime,omitempty" json:"effectiveDateTime,omitempty"`
	PerformedDateTime *string `bson:"performedDateTime,omitempty" json:"performedDateTime,omitempty"`
	RecordedDate      *string `bson:"recordedDate,omitempty" json:"recordedDate,omitempty"`
	EffectivePeriod   *Period `bson:"effectivePeriod,omitempty" json:"effectivePeriod,omitempty"`
	Period            *Period `bson:"period,omitempty" json:"period,omitempty"`
}

func (f *DateFilter) apply(fhirData []byte) bool {
	bundle, err := models.UnmarshalBundle(fhirData)
	check(err)

	for _, e := range bundle.Entry {
		var r DateTimeResource
		err = json.Unmarshal(e.Resource, &r)
		check(err)

		if *r.Type == "Patient" {
			return true
		}

		value := element(r)

		switch value.(type) {
		case nil:
			return false
		case *string:
			if f.applyDateTime(value.(*string)) {
				return true
			}
		case *Period:
			if f.applyPeriod(value.(*Period)) {
				return true
			}
		}

	}
	return false
}

func element(r DateTimeResource) interface{} {
	switch {
	case r.EffectiveDateTime != nil:
		return r.EffectiveDateTime
	case r.PerformedDateTime != nil:
		return r.PerformedDateTime
	case r.RecordedDate != nil:
		return r.RecordedDate
	case r.EffectivePeriod != nil:
		return r.EffectivePeriod
	case r.Period != nil:
		return r.Period
	}
	return nil
}

func (f *DateFilter) applyDateTime(dateTime *string) bool {
	dt, err := time.Parse(time.RFC3339, *dateTime)
	if err != nil {
		check(err)
		return false
	}

	return f.Comparator(dt) || (f.Inclusive && f.Date.Equal(dt))
}

func (f *DateFilter) applyPeriod(period *Period) bool {
	return f.applyDateTime(period.Start) || f.applyDateTime(period.End)
}

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
	var comp func(t time.Time) bool
	incl := false

	switch config.Comparator {
	case "", "=":
		comp = config.Value.Equal
		incl = true
	case ">":
		comp = config.Value.Before
	case ">=":
		comp = config.Value.Before
		incl = true
	case "<":
		comp = config.Value.After
	case "<=":
		comp = config.Value.After
		incl = true
	}

	return &DateFilter{Date: *config.Value, Comparator: comp, Inclusive: incl}
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
	AuthoredOn        *string `bson:"authoredOn,omitempty" json:"authoredOn,omitempty"`
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

		switch value := element(r).(type) {
		case nil:
			return false
		case *string:
			if f.applyDateTime(value) {
				return true
			}
		case *Period:
			if f.applyPeriod(value) {
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
	case r.AuthoredOn != nil:
		return r.AuthoredOn
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

	// compare equal date parts if comparison includes the date
	return (f.Inclusive && f.Date.Year() == dt.Year() && f.Date.YearDay() == dt.YearDay()) || f.Comparator(dt)
}

func (f *DateFilter) applyPeriod(period *Period) bool {
	return f.applyDateTime(period.Start) || f.applyDateTime(period.End)
}

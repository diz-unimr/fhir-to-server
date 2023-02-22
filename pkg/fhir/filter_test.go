package fhir

import (
	"fhir-to-server/pkg/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var loc *time.Location

func TestApply(t *testing.T) {
	loc, _ = time.LoadLocation("Europe/Berlin")

	for scenario, fn := range map[string]func(t *testing.T){
		"Patient valid": testApplyPassesPatient,
		"DiagnosticReport.effectiveDatetime valid": testApplyFilterValidDiagnosticReport,
		"Procedure.recordedDate skip":              testApplyFilterSkipProcedureRecorded,
		"Encounter.Period valid":                   testApplyFilterValidEncounterPeriod,
		"Condition.period valid inclusive":         testApplyFilterValidConditionInclusive,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func testApplyFilterValidDiagnosticReport(t *testing.T) {
	t.Parallel()
	conf := config.DateConfig{Value: createTime("2018-03-01"), Comparator: ">"}
	f := NewDateFilter(conf)
	testBundle := []byte(`
{
 "resourceType": "Bundle",
 "type": "batch",
 "entry": [
   {
     "resource": {
       "resourceType": "DiagnosticReport",
	   "effectiveDateTime": "2023-02-20T12:52:00+01:00"
     }
   }
 ]
}
`)

	passed := f.apply(testBundle)
	assert.Truef(t, passed, "Expected bundle (DiagnosticReport.effectiveDatetime: 2023-02-20T12:52:00+01:00) "+
		"to pass the filter (Date: %s, Comparator: %s) but it didn't", conf.Value, conf.Comparator)
}

func testApplyFilterSkipProcedureRecorded(t *testing.T) {
	t.Parallel()
	conf := config.DateConfig{Value: createTime("2018-03-01"), Comparator: "<"}
	f := NewDateFilter(conf)
	testBundle := []byte(`
{
 "resourceType": "Bundle",
 "type": "batch",
 "entry": [
   {
     "resource": {
       "resourceType": "Procedure",
       "recordedDate": "2023-02-20T12:52:00+01:00"
     }
   }
 ]
}
`)

	passed := f.apply(testBundle)
	assert.Falsef(t, passed, "Expected bundle (DiagnosticReport.effectiveDatetime: 2023-02-20T12:52:00+01:00) "+
		"to be skipped by the filter (Date: %s, Comparator: %s) but it wasn't", conf.Value, conf.Comparator)
}

func testApplyPassesPatient(t *testing.T) {
	t.Parallel()
	conf := config.DateConfig{Value: createTime("2018-03-01"), Comparator: ">"}
	f := NewDateFilter(conf)
	testBundle := []byte(`
{
 "resourceType": "Bundle",
 "type": "batch",
 "entry": [
   {
     "resource": {
       "resourceType": "Patient"
     }
   }
 ]
}
`)

	passed := f.apply(testBundle)
	assert.Truef(t, passed, "Expected bundle (Patient resource) "+
		"to pass the filter (Date: %s, Comparator: %s) but it didn't", conf.Value, conf.Comparator)
}

func testApplyFilterValidEncounterPeriod(t *testing.T) {
	t.Parallel()
	conf := config.DateConfig{Value: createTime("2018-03-01"), Comparator: ">"}
	f := NewDateFilter(conf)
	testBundle := []byte(`
{
 "resourceType": "Bundle",
 "type": "batch",
 "entry": [
   {
     "resource": {
       "resourceType": "Encounter",
	   "period": {
	     "start": "2018-02-28T01:00:00+01:00",
         "end": "2018-03-10T20:00:00+01:00"
	   }
     }
   }
 ]
}
`)

	passed := f.apply(testBundle)
	assert.Truef(t, passed, "Expected bundle (Encounter.period.start: 2018-02-28T01:00:00+01:00, Encounter.period.end: 2018-03-10T20:00:00+01:00})) "+
		"to pass the filter (Date: %s, Comparator: %s) but it didn't", conf.Value, conf.Comparator)
}

func testApplyFilterValidConditionInclusive(t *testing.T) {
	t.Parallel()
	conf := config.DateConfig{Value: createTime("2018-03-01"), Comparator: "="}
	f := NewDateFilter(conf)
	testBundle := []byte(`
{
 "resourceType": "Bundle",
 "type": "batch",
 "entry": [
   {
     "resource": {
       "resourceType": "Condition",
	   "recordedDate": "2018-03-01T00:00:00+01:00"
     }
   }
 ]
}
`)

	passed := f.apply(testBundle)
	assert.Truef(t, passed, "Expected bundle (Condition.recordedDate: 2018-03-01T00:00:00+01:00) "+
		"to pass the filter (Date: %s, Comparator: %s) but it didn't", conf.Value, conf.Comparator)
}

func createTime(value string) *time.Time {
	t, _ := time.ParseInLocation("2006-01-02", value, loc)
	return &t
}

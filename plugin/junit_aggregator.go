package plugin

import (
	"encoding/xml"
	"github.com/sirupsen/logrus"
)

type JunitAggregator struct {
	ReportsDir  string
	ReportsName string
	Includes    string
	DbCredentials
}

type JunitAggregatorData struct {
	ResultBasicInfo
	XMLName   xml.Name `xml:"testsuite"`
	Name      string   `xml:"name,attr"`
	Tests     int      `xml:"tests,attr"`
	Skipped   int      `xml:"skipped,attr"`
	Failures  int      `xml:"failures,attr"`
	Errors    int      `xml:"errors,attr"`
	Timestamp string   `xml:"timestamp,attr,omitempty"`
	Hostname  string   `xml:"hostname,attr,omitempty"`
	Time      float64  `xml:"time,attr"`
	Version   string   `xml:"version,attr,omitempty"`
	Schema    string   `xml:"xsi:noNamespaceSchemaLocation,attr,omitempty"`
}

func GetNewJunitAggregator(
	reportsDir, reportsName, includes, dbUrl, dbToken, dbOrg, dbBucket string) *JunitAggregator {
	return &JunitAggregator{
		ReportsDir:  reportsDir,
		ReportsName: reportsName,
		Includes:    includes,
		DbCredentials: DbCredentials{
			InfluxDBURL:   dbUrl,
			InfluxDBToken: dbToken,
			Organization:  dbOrg,
			Bucket:        dbBucket,
		},
	}
}

func (j *JunitAggregator) Aggregate(groupName string) error {
	logrus.Println("JunitAggregator Aggregator Aggregate")
	err := Aggregate[JunitAggregatorData](j.ReportsDir, j.Includes,
		j.DbCredentials.InfluxDBURL, j.DbCredentials.InfluxDBToken,
		j.DbCredentials.Organization, j.DbCredentials.Bucket, JunitTool, groupName,
		CalculateJunitAggregate, GetJunitDataMaps)
	return err
}

func CalculateJunitAggregate(junitAggregatorList []JunitAggregatorData) JunitAggregatorData {
	aggregatorData := JunitAggregatorData{}

	for _, junitAggregatorData := range junitAggregatorList {
		aggregatorData.Tests += junitAggregatorData.Tests
		aggregatorData.Skipped += junitAggregatorData.Skipped
		aggregatorData.Failures += junitAggregatorData.Failures
		aggregatorData.Errors += junitAggregatorData.Errors
	}

	return aggregatorData
}

func GetJunitDataMaps(pipelineId, buildNumber string,
	aggregateData JunitAggregatorData) (map[string]string, map[string]interface{}) {
	tagsMap := map[string]string{
		"pipeline_id": pipelineId,
		"build_id":    buildNumber,
		"name":        aggregateData.Name,
		"type":        aggregateData.Type,
		"status":      aggregateData.Status,
	}

	fieldsMap := map[string]interface{}{
		"tests":    aggregateData.Tests,
		"skipped":  aggregateData.Skipped,
		"failures": aggregateData.Failures,
		"errors":   aggregateData.Errors,
	}

	return tagsMap, fieldsMap
}

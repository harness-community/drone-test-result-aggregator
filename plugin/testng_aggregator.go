package plugin

import (
	"encoding/xml"
	"fmt"
)

type TestNgAggregator struct {
	ReportsDir  string
	ReportsName string
	Includes    string
	DbCredentials
}

type TestNGResults struct {
	XMLName xml.Name `xml:"testng-results"`
	Ignored int      `xml:"ignored,attr"`
	Total   int      `xml:"total,attr"`
	Passed  int      `xml:"passed,attr"`
	Failed  int      `xml:"failed,attr"`
	Skipped int      `xml:"skipped,attr"`
}

func GetNewTestNgAggregator(
	reportsDir, reportsName, includes, dbUrl, dbToken, dbOrg, dbBucket string) *TestNgAggregator {
	return &TestNgAggregator{
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

func (t *TestNgAggregator) Aggregate() error {
	fmt.Println("TestNgAggregator Aggregator Aggregate")

	err := Aggregate[TestNGResults](t.ReportsDir, t.Includes,
		t.DbCredentials.InfluxDBURL, t.DbCredentials.InfluxDBToken,
		t.DbCredentials.Organization, t.DbCredentials.Bucket, TestNgTool,
		CalculateTestNgAggregate, GetTestNgDataMaps)
	return err
}

func CalculateTestNgAggregate(testNgAggregatorList []TestNGResults) TestNGResults {
	aggregatorData := TestNGResults{}

	for _, testNgAggregatorData := range testNgAggregatorList {
		aggregatorData.Ignored += testNgAggregatorData.Ignored
		aggregatorData.Total += testNgAggregatorData.Total
		aggregatorData.Passed += testNgAggregatorData.Passed
		aggregatorData.Failed += testNgAggregatorData.Failed
		aggregatorData.Skipped += testNgAggregatorData.Skipped
	}

	return aggregatorData
}

func GetTestNgDataMaps(pipelineId, buildNumber string,
	aggregateData TestNGResults) (map[string]string, map[string]interface{}) {
	tagsMap := map[string]string{
		"pipelineId": pipelineId,
		"buildId":    buildNumber,
	}
	fieldsMap := map[string]interface{}{
		"ignored": aggregateData.Ignored,
		"total":   aggregateData.Total,
		"passed":  aggregateData.Passed,
		"failed":  aggregateData.Failed,
		"skipped": aggregateData.Skipped,
	}
	return tagsMap, fieldsMap
}

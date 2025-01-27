package plugin

import (
	"encoding/xml"
	"fmt"
	"strings"
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

	reportsRootDir := t.ReportsDir
	patterns := strings.Split(t.Includes, ",")

	testNgAggregatorList, err := GetXmlReportData[TestNGResults](reportsRootDir, patterns)
	if err != nil {
		fmt.Println("Error getting xml report data: ", err.Error())
		return err
	}

	totalAggregate := t.calculateAggregate(testNgAggregatorList)
	fmt.Println("Total Aggregate: ", totalAggregate)

	s, err := ToJsonStringFromStruct[TestNGResults](totalAggregate)
	if err != nil {
		fmt.Println("Error converting struct to json string: ", err.Error())
		return err
	}
	fmt.Println("Total Aggregate in json: ", s)

	pipelineId, buildNumber, err := GetPipelineInfo()
	if err != nil {
		fmt.Println("Error getting pipeline info: ", err.Error())
		return err
	}

	err = t.PersistToInfluxDb(pipelineId, buildNumber, totalAggregate)
	return err
}

func (t *TestNgAggregator) PersistToInfluxDb(pipelineId, buildNumber string, aggregateData TestNGResults) error {
	return nil
}

func (t *TestNgAggregator) calculateAggregate(testNgAggregatorList []TestNGResults) TestNGResults {

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

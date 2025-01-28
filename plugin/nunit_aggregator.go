package plugin

import (
	"encoding/xml"
	"fmt"
)

type NunitAggregator struct {
	ReportsDir  string
	ReportsName string
	Includes    string
	DbCredentials
}

type TestRun struct {
	XMLName       xml.Name      `xml:"TestRun"`
	ResultSummary ResultSummary `xml:"ResultSummary"`
}

type ResultSummary struct {
	Outcome  string   `xml:"outcome,attr"`
	Counters Counters `xml:"Counters"`
}

type Counters struct {
	Total               int `xml:"total,attr"`
	Executed            int `xml:"executed,attr"`
	Passed              int `xml:"passed,attr"`
	Failed              int `xml:"failed,attr"`
	Error               int `xml:"error,attr"`
	Timeout             int `xml:"timeout,attr"`
	Aborted             int `xml:"aborted,attr"`
	Inconclusive        int `xml:"inconclusive,attr"`
	PassedButRunAborted int `xml:"passedButRunAborted,attr"`
	NotRunnable         int `xml:"notRunnable,attr"`
	NotExecuted         int `xml:"notExecuted,attr"`
	Disconnected        int `xml:"disconnected,attr"`
	Warning             int `xml:"warning,attr"`
	Completed           int `xml:"completed,attr"`
	InProgress          int `xml:"inProgress,attr"`
	Pending             int `xml:"pending,attr"`
}

func GetNewNunitAggregator(
	reportsDir, reportsName, includes, dbUrl, dbToken, dbOrg, dbBucket string) *NunitAggregator {
	return &NunitAggregator{
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

func (n *NunitAggregator) Aggregate() error {

	fmt.Println("Nunit Aggregator Aggregate")

	err := Aggregate[TestRun](n.ReportsDir, n.Includes,
		n.DbCredentials.InfluxDBURL, n.DbCredentials.InfluxDBToken,
		n.DbCredentials.Organization, n.DbCredentials.Bucket, NunitTool,
		CalculateNugetAggregate, GetNunitDataMaps)

	return err
}

func CalculateNugetAggregate(nunitAggregatorList []TestRun) TestRun {
	aggregatorData := TestRun{}

	for _, nunitAggregatorData := range nunitAggregatorList {
		aggregatorData.ResultSummary.Outcome = nunitAggregatorData.ResultSummary.Outcome
		aggregatorData.ResultSummary.Counters.Total += nunitAggregatorData.ResultSummary.Counters.Total
		aggregatorData.ResultSummary.Counters.Executed += nunitAggregatorData.ResultSummary.Counters.Executed
		aggregatorData.ResultSummary.Counters.Passed += nunitAggregatorData.ResultSummary.Counters.Passed
		aggregatorData.ResultSummary.Counters.Failed += nunitAggregatorData.ResultSummary.Counters.Failed
		aggregatorData.ResultSummary.Counters.Error += nunitAggregatorData.ResultSummary.Counters.Error
		aggregatorData.ResultSummary.Counters.Timeout += nunitAggregatorData.ResultSummary.Counters.Timeout
		aggregatorData.ResultSummary.Counters.Aborted += nunitAggregatorData.ResultSummary.Counters.Aborted
		aggregatorData.ResultSummary.Counters.Inconclusive += nunitAggregatorData.ResultSummary.Counters.Inconclusive
		aggregatorData.ResultSummary.Counters.PassedButRunAborted += nunitAggregatorData.ResultSummary.Counters.PassedButRunAborted
		aggregatorData.ResultSummary.Counters.NotRunnable += nunitAggregatorData.ResultSummary.Counters.NotRunnable
		aggregatorData.ResultSummary.Counters.NotExecuted += nunitAggregatorData.ResultSummary.Counters.NotExecuted
		aggregatorData.ResultSummary.Counters.Disconnected += nunitAggregatorData.ResultSummary.Counters.Disconnected
		aggregatorData.ResultSummary.Counters.Warning += nunitAggregatorData.ResultSummary.Counters.Warning
		aggregatorData.ResultSummary.Counters.Completed += nunitAggregatorData.ResultSummary.Counters.Completed
		aggregatorData.ResultSummary.Counters.InProgress += nunitAggregatorData.ResultSummary.Counters.InProgress
		aggregatorData.ResultSummary.Counters.Pending += nunitAggregatorData.ResultSummary.Counters.Pending
	}
	return aggregatorData
}

func GetNunitDataMaps(pipelineId, buildNumber string,
	aggregateData TestRun) (map[string]string, map[string]interface{}) {
	tagsMap := map[string]string{
		"pipelineId": pipelineId,
		"buildId":    buildNumber,
	}
	fieldsMap := map[string]interface{}{
		"outcome":             aggregateData.ResultSummary.Outcome,
		"total":               aggregateData.ResultSummary.Counters.Total,
		"executed":            aggregateData.ResultSummary.Counters.Executed,
		"passed":              aggregateData.ResultSummary.Counters.Passed,
		"failed":              aggregateData.ResultSummary.Counters.Failed,
		"error":               aggregateData.ResultSummary.Counters.Error,
		"timeout":             aggregateData.ResultSummary.Counters.Timeout,
		"aborted":             aggregateData.ResultSummary.Counters.Aborted,
		"inconclusive":        aggregateData.ResultSummary.Counters.Inconclusive,
		"passedButRunAborted": aggregateData.ResultSummary.Counters.PassedButRunAborted,
		"notRunnable":         aggregateData.ResultSummary.Counters.NotRunnable,
		"notExecuted":         aggregateData.ResultSummary.Counters.NotExecuted,
		"disconnected":        aggregateData.ResultSummary.Counters.Disconnected,
		"warning":             aggregateData.ResultSummary.Counters.Warning,
		"completed":           aggregateData.ResultSummary.Counters.Completed,
		"inProgress":          aggregateData.ResultSummary.Counters.InProgress,
		"pending":             aggregateData.ResultSummary.Counters.Pending,
	}
	return tagsMap, fieldsMap
}

package plugin

import (
	"encoding/xml"
	"fmt"
	"strings"
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

	reportsRootDir := n.ReportsDir
	patterns := strings.Split(n.Includes, ",")

	nunitAggregatorDataList, err := GetXmlReportData[TestRun](reportsRootDir, patterns)
	if err != nil {
		fmt.Println("Error getting xml report data: ", err.Error())
		return err
	}

	totalAggregate := n.calculateAggregate(nunitAggregatorDataList)
	fmt.Println("Total Aggregate: ", totalAggregate)

	s, err := ToJsonStringFromStruct[ResultSummary](totalAggregate)
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

	err = n.PersistToInfluxDb(pipelineId, buildNumber, totalAggregate)
	if err != nil {
		fmt.Println("Error persisting to influxdb: ", err.Error())
	}
	return err
}

func (n *NunitAggregator) PersistToInfluxDb(pipelineId, buildNumber string, aggregateData ResultSummary) error {
	tagsMap := map[string]string{
		"pipelineId": pipelineId,
		"buildId":    buildNumber,
	}
	fieldsMap := map[string]interface{}{
		"outcome":             aggregateData.Outcome,
		"total":               aggregateData.Counters.Total,
		"executed":            aggregateData.Counters.Executed,
		"passed":              aggregateData.Counters.Passed,
		"failed":              aggregateData.Counters.Failed,
		"error":               aggregateData.Counters.Error,
		"timeout":             aggregateData.Counters.Timeout,
		"aborted":             aggregateData.Counters.Aborted,
		"inconclusive":        aggregateData.Counters.Inconclusive,
		"passedButRunAborted": aggregateData.Counters.PassedButRunAborted,
		"notRunnable":         aggregateData.Counters.NotRunnable,
		"notExecuted":         aggregateData.Counters.NotExecuted,
		"disconnected":        aggregateData.Counters.Disconnected,
		"warning":             aggregateData.Counters.Warning,
		"completed":           aggregateData.Counters.Completed,
		"inProgress":          aggregateData.Counters.InProgress,
		"pending":             aggregateData.Counters.Pending,
	}

	err := PersistToInfluxDb(n.DbCredentials.InfluxDBURL, n.DbCredentials.InfluxDBToken,
		n.DbCredentials.Organization, n.DbCredentials.Bucket, NunitTool,
		tagsMap, fieldsMap)

	return err
}
func (n *NunitAggregator) calculateAggregate(nunitAggregatorDataList []TestRun) ResultSummary {

	aggregatorData := ResultSummary{}

	for _, nunitAggregatorData := range nunitAggregatorDataList {
		aggregatorData.Outcome = nunitAggregatorData.ResultSummary.Outcome
		aggregatorData.Counters.Total += nunitAggregatorData.ResultSummary.Counters.Total
		aggregatorData.Counters.Executed += nunitAggregatorData.ResultSummary.Counters.Executed
		aggregatorData.Counters.Passed += nunitAggregatorData.ResultSummary.Counters.Passed
		aggregatorData.Counters.Failed += nunitAggregatorData.ResultSummary.Counters.Failed
		aggregatorData.Counters.Error += nunitAggregatorData.ResultSummary.Counters.Error
		aggregatorData.Counters.Timeout += nunitAggregatorData.ResultSummary.Counters.Timeout
		aggregatorData.Counters.Aborted += nunitAggregatorData.ResultSummary.Counters.Aborted
		aggregatorData.Counters.Inconclusive += nunitAggregatorData.ResultSummary.Counters.Inconclusive
		aggregatorData.Counters.PassedButRunAborted += nunitAggregatorData.ResultSummary.Counters.PassedButRunAborted
		aggregatorData.Counters.NotRunnable += nunitAggregatorData.ResultSummary.Counters.NotRunnable
		aggregatorData.Counters.NotExecuted += nunitAggregatorData.ResultSummary.Counters.NotExecuted
		aggregatorData.Counters.Disconnected += nunitAggregatorData.ResultSummary.Counters.Disconnected
		aggregatorData.Counters.Warning += nunitAggregatorData.ResultSummary.Counters.Warning
		aggregatorData.Counters.Completed += nunitAggregatorData.ResultSummary.Counters.Completed
		aggregatorData.Counters.InProgress += nunitAggregatorData.ResultSummary.Counters.InProgress
		aggregatorData.Counters.Pending += nunitAggregatorData.ResultSummary.Counters.Pending
	}
	return aggregatorData
}

func (n *NunitAggregator) GetDbUrl() string {
	return n.DbCredentials.InfluxDBURL
}

func (n *NunitAggregator) GetDbToken() string {
	return n.DbCredentials.InfluxDBToken
}

func (n *NunitAggregator) GetDbOrganization() string {
	return n.DbCredentials.Organization
}

func (n *NunitAggregator) GetDbBucket() string {
	return n.DbCredentials.Bucket
}

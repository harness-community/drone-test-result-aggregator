package plugin

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type NunitAggregator struct {
	ReportsDir  string
	ReportsName string
	Includes    string
	DbCredentials
}

type TestRunSummary struct {
	TotalCases   int    `xml:"total,attr"`
	TotalPassed  int    `xml:"passed,attr"`
	TotalFailed  int    `xml:"failed,attr"`
	TotalSkipped int    `xml:"skipped,attr"`
	Result       string `xml:"result,attr"`
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

func (n *NunitAggregator) Aggregate(groupName string) error {
	logrus.Println("NUnit Aggregator Aggregate (Using <test-run> Summary)")

	tagsMap, fieldsMap, err := Aggregate[TestRunSummary](n.ReportsDir, n.Includes,
		n.DbCredentials.InfluxDBURL, n.DbCredentials.InfluxDBToken,
		n.DbCredentials.Organization, n.DbCredentials.Bucket, NunitTool, groupName,
		CalculateNunitAggregate, GetNunitDataMaps, ShowNunitStats)
	_, _ = tagsMap, fieldsMap
	return err
}

func CalculateNunitAggregate(reports []TestRunSummary) TestRunSummary {
	fmt.Println("CalculateNunitAggregate - Using only <test-run> summary")

	totalCases, totalPassed, totalFailed, totalSkipped := 0, 0, 0, 0

	for _, report := range reports {
		totalCases += report.TotalCases
		totalPassed += report.TotalPassed
		totalFailed += report.TotalFailed
		totalSkipped += report.TotalSkipped
	}

	fmt.Printf("Summary -> Total: %d, Passed: %d, Failed: %d, Skipped: %d\n",
		totalCases, totalPassed, totalFailed, totalSkipped)

	return TestRunSummary{
		TotalCases:   totalCases,
		TotalPassed:  totalPassed,
		TotalFailed:  totalFailed,
		TotalSkipped: totalSkipped,
		Result:       "Aggregated",
	}
}

func GetNunitDataMaps(pipelineId, buildNumber string, aggregateData TestRunSummary) (map[string]string, map[string]interface{}) {
	fmt.Println("GetNunitDataMaps - Using only <test-run> summary")

	tags := map[string]string{
		"pipelineId": pipelineId,
		"buildId":    buildNumber,
	}

	fields := map[string]interface{}{
		"total_cases":   aggregateData.TotalCases,
		"total_passed":  aggregateData.TotalPassed,
		"total_failed":  aggregateData.TotalFailed,
		"total_skipped": aggregateData.TotalSkipped,
	}

	return tags, fields
}

func ShowNunitStats(tags map[string]string, fields map[string]interface{}) error {
	fmt.Println("")
	fmt.Println("====================================================================")
	fmt.Println("NUnit Test Run Summary")
	fmt.Printf("Pipeline ID: %s, Build ID: %s \n", tags["pipelineId"], tags["buildId"])
	fmt.Println("====================================================================")
	fmt.Println("📁 Total Cases: ", fields["total_cases"])
	fmt.Println("✅ Total Passed: ", fields["total_passed"])
	fmt.Println("❌ Total Failed: ", fields["total_failed"])
	fmt.Println("⏸️ Total Skipped: ", fields["total_skipped"])
	return nil
}

package plugin

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
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
	if err != nil {
		return fmt.Errorf("failed to aggregate NUnit test results: %w", err)
	}

	err = ExportNunitOutputVars(tagsMap, fieldsMap)
	if err != nil {
		logrus.Println("Error exporting Nunit output variables", err)
		return err
	}
	return nil
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

func ExportNunitOutputVars(tags map[string]string, fields map[string]interface{}) error {
	outputVarsMap := map[string]interface{}{
		"TOTAL_CASES":   fields["total_cases"],
		"TOTAL_PASSED":  fields["total_passed"],
		"TOTAL_FAILED":  fields["total_failed"],
		"TOTAL_SKIPPED": fields["total_skipped"],
	}
	for key, value := range outputVarsMap {
		err := WriteToEnvVariable(key, fmt.Sprintf("%v", value))
		if err != nil {
			logrus.Errorf("Error writing %s to env variable: %v", key, err)
			return err
		}
	}
	return nil
}

func ShowNunitStats(tags map[string]string, fields map[string]interface{}) error {
	border := "=================================="
	separator := "----------------------------------"

	table := []string{
		border,
		"  NUnit Test Run Summary",
		border,
		fmt.Sprintf(" Pipeline ID : %-40s", tags["pipelineId"]),
		fmt.Sprintf(" Build ID : %-40s", tags["buildId"]),
		border,
		fmt.Sprintf("| %-16s | %-10s |", "Test Category", "Count   "),
		separator,
		fmt.Sprintf("| 📁 Total Cases   | %10.0f |", float64(fields["total_cases"].(int))),
		fmt.Sprintf("| ✅ Total Passed  | %10.0f |", float64(fields["total_passed"].(int))),
		fmt.Sprintf("| ❌ Total Failed  | %10.0f |", float64(fields["total_failed"].(int))),
		fmt.Sprintf("| ⏸️ Total Skipped | %10.0f |", float64(fields["total_skipped"].(int))),
		border,
	}

	fmt.Println(strings.Join(table, "\n"))
	return nil
}

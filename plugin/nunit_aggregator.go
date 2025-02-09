package plugin

import (
	"encoding/csv"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
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

	err = WriteNunitMetricsCsvData(TestResultsDataFileCsv, tagsMap, fieldsMap)
	if err != nil {
		logrus.Errorf("Error writing NUnit metrics to CSV file: %v", err)
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

// ShowNunitStats displays NUnit test summary in a well-formatted table.
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

func WriteNunitMetricsCsvData(csvFileName string, tagsMap map[string]string, fieldsMap map[string]interface{}) error {
	file, err := os.Create(csvFileName)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	var csvBuffer strings.Builder
	bufferWriter := csv.NewWriter(&csvBuffer)

	header := []string{"Pipeline ID", "Build ID", "Test Category", "Count"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header to file: %w", err)
	}
	if err := bufferWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header to buffer: %w", err)
	}

	nunitData := [][]string{
		{tagsMap["pipelineId"], tagsMap["buildId"], "Total Cases", fmt.Sprintf("%d", fieldsMap["total_cases"].(int))},
		{tagsMap["pipelineId"], tagsMap["buildId"], "Total Passed", fmt.Sprintf("%d", fieldsMap["total_passed"].(int))},
		{tagsMap["pipelineId"], tagsMap["buildId"], "Total Failed", fmt.Sprintf("%d", fieldsMap["total_failed"].(int))},
		{tagsMap["pipelineId"], tagsMap["buildId"], "Total Skipped", fmt.Sprintf("%d", fieldsMap["total_skipped"].(int))},
	}

	for _, row := range nunitData {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row to file: %w", err)
		}
		if err := bufferWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row to buffer: %w", err)
		}
	}

	writer.Flush()
	bufferWriter.Flush()

	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing CSV writer to file: %w", err)
	}

	err = WriteToEnvVariable(TestResultsDataFile, csvFileName)
	if err != nil {
		logrus.Errorf("Error writing CSV file path to env variable: %v", err)
		return err
	}

	logrus.Infof("NUnit test metrics exported to %s and stored in NUNIT_METRICS_CSV_FILE env variable", csvFileName)
	return nil
}

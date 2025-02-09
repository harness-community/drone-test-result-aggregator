package plugin

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
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

type TestNGReport struct {
	XMLName           xml.Name `xml:"testng-results"`
	Suites            []Suite  `xml:"suite"`
	AggregatedResults Results
}

type Suite struct {
	Name     string  `xml:"name,attr"`
	Duration string  `xml:"duration-ms,attr"`
	Groups   []Group `xml:"groups>group"`
	Classes  []Class `xml:"test>class"`
}

type Group struct {
	Name    string   `xml:"name,attr"`
	Methods []Method `xml:"method"`
}

type Method struct {
	Name      string `xml:"name,attr"`
	Signature string `xml:"signature,attr"`
	ClassName string `xml:"class,attr"`
}

type Class struct {
	Name  string `xml:"name,attr"`
	Tests []Test `xml:"test-method"`
}

type Test struct {
	Name        string `xml:"name,attr"`
	Status      string `xml:"status,attr"`
	DurationMS  string `xml:"duration-ms,attr"`
	IsConfig    bool   `xml:"is-config,attr"`
	Description string `xml:"description,attr"`
	Exception   string `xml:"exception>short-stacktrace"`
}

type Results struct {
	Total      int
	Failures   int
	Skipped    int
	DurationMS float64
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

func (t *TestNgAggregator) Aggregate(groupName string) error {
	logrus.Println("TestNgAggregator Aggregator Aggregate")

	tagsMap, fieldsMap, err := Aggregate[TestNGReport](t.ReportsDir, t.Includes,
		t.DbCredentials.InfluxDBURL, t.DbCredentials.InfluxDBToken,
		t.DbCredentials.Organization, t.DbCredentials.Bucket, TestNgTool, groupName,
		CalculateTestNgAggregate, GetTestNgDataMaps, ShowTestNgStats)
	if err != nil {
		logrus.Errorf("Error aggregating TestNG results: %v", err)
		return err
	}

	err = WriteTestNgMetricsCsvData(TestResultsDataFileCsv, tagsMap, fieldsMap)
	if err != nil {
		logrus.Errorf("Error writing TestNG metrics to CSV file: %v", err)
		return err
	}
	return nil

}

func CalculateTestNgAggregate(testNgAggregatorList []TestNGReport) TestNGReport {
	aggregatorData := TestNGReport{}
	var totalTests, totalFailures, totalSkipped int
	var totalDuration float64

	for _, report := range testNgAggregatorList {
		for _, suite := range report.Suites {
			suiteResults, _, _ := aggregateSuiteResults(suite)

			totalTests += suiteResults.Total
			totalFailures += suiteResults.Failures
			totalSkipped += suiteResults.Skipped
			totalDuration += suiteResults.DurationMS
		}
	}

	aggregatorData.AggregatedResults = Results{
		Total:      totalTests,
		Failures:   totalFailures,
		Skipped:    totalSkipped,
		DurationMS: totalDuration,
	}

	return aggregatorData
}

func GetTestNgDataMaps(pipelineId, buildNumber string,
	aggregateData TestNGReport) (map[string]string, map[string]interface{}) {

	tags := map[string]string{
		"pipelineId": pipelineId,
		"buildId":    buildNumber,
	}

	fields := map[string]interface{}{
		"total_cases":   aggregateData.AggregatedResults.Total,
		"total_failed":  aggregateData.AggregatedResults.Failures,
		"total_skipped": aggregateData.AggregatedResults.Skipped,
		"duration_ms":   aggregateData.AggregatedResults.DurationMS,
	}

	return tags, fields
}

func aggregateSuiteResults(suite Suite) (Results, []string, []string) {
	results := Results{}
	var failedTests []string
	var skippedTests []string

	for _, class := range suite.Classes {
		classResults, failed, skipped := aggregateClassResults(class)
		results.Total += classResults.Total
		results.Failures += classResults.Failures
		results.Skipped += classResults.Skipped
		results.DurationMS += classResults.DurationMS

		failedTests = append(failedTests, failed...)
		skippedTests = append(skippedTests, skipped...)
	}

	return results, failedTests, skippedTests
}

func aggregateClassResults(class Class) (Results, []string, []string) {
	results := Results{}
	var failedTests []string
	var skippedTests []string

	for _, test := range class.Tests {
		results.Total++
		if test.Status == "FAIL" {
			results.Failures++
			failedTests = append(failedTests, test.Name)
		} else if test.Status == "SKIP" {
			results.Skipped++
			skippedTests = append(skippedTests, test.Name)
		}

		duration, err := strconv.ParseFloat(test.DurationMS, 64)
		if err != nil {
			logrus.Warnf("Invalid or missing DurationMS for test '%s': %v", test.Name, err)
			continue
		}
		results.DurationMS += duration
	}

	return results, failedTests, skippedTests
}

func WriteTestNgMetricsCsvData(csvFileName string, tagsMap map[string]string, fieldsMap map[string]interface{}) error {
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

	testNgData := [][]string{
		{tagsMap["pipeline_id"], tagsMap["build_number"], "Total Cases", fmt.Sprintf("%d", fieldsMap["total_cases"].(int))},
		{tagsMap["pipeline_id"], tagsMap["build_number"], "Total Failed", fmt.Sprintf("%d", fieldsMap["total_failed"].(int))},
		{tagsMap["pipeline_id"], tagsMap["build_number"], "Total Skipped", fmt.Sprintf("%d", fieldsMap["total_skipped"].(int))},
		{tagsMap["pipeline_id"], tagsMap["build_number"], "Total Duration (ms)", fmt.Sprintf("%.2f", fieldsMap["duration_ms"].(float64))},
	}

	for _, row := range testNgData {
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

	logrus.Infof("TestNG test metrics exported to %s and stored in TESTNG_METRICS_CSV_FILE env variable", csvFileName)
	return nil
}

func ShowTestNgStats(tagsMap map[string]string, fieldsMap map[string]interface{}) error {
	border := "==================================================================="
	separator := "-------------------------------------------------------------------"

	table := []string{
		border,
		"  TestNG Test Run Summary",
		border,
		fmt.Sprintf("  Pipeline ID      : %-40s", tagsMap["pipelineId"]),
		fmt.Sprintf("  Build ID         : %-40s", tagsMap["buildId"]),
		border,
		fmt.Sprintf("| %-22s | %-10s |", "Test Category", "Count"),
		separator,
		fmt.Sprintf("| 📁 Total Cases     | %10.0f    |", float64(fieldsMap["total_cases"].(int))),
		fmt.Sprintf("| ❌ Total Failed     | %10.0f    |", float64(fieldsMap["total_failed"].(int))),
		fmt.Sprintf("| ⏸️ Total Skipped   | %10.0f    |", float64(fieldsMap["total_skipped"].(int))),
		fmt.Sprintf("| ⏱️ Total Duration  | %10.2f ms |", fieldsMap["duration_ms"].(float64)),
		border,
	}

	fmt.Println(strings.Join(table, "\n"))
	return nil
}

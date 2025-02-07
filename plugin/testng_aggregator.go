package plugin

import (
	"encoding/xml"
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
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
	_, _ = tagsMap, fieldsMap
	return err
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

func ShowTestNgStats(tagsMap map[string]string, fieldsMap map[string]interface{}) error {
	fmt.Println("")
	fmt.Println("====================================================================")
	fmt.Println("TestNG Test Run Summary")
	fmt.Printf("Pipeline ID: %s, Build ID: %s \n", tagsMap["pipeline_id"], tagsMap["build_number"])
	fmt.Println("====================================================================")
	fmt.Println("📁 Total Cases:   ", fieldsMap["total_cases"])
	fmt.Println("❌ Total Failed:  ", fieldsMap["total_failed"])
	fmt.Println("⏸️ Total Skipped: ", fieldsMap["total_skipped"])
	fmt.Println("⏱️ Total Duration (ms): ", fieldsMap["duration_ms"])
	fmt.Println("====================================================================")
	return nil
}

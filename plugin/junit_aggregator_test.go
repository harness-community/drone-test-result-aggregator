package plugin

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

type MockJunitTestStats struct {
	TestCount    int
	FailCount    int
	PassCount    int
	SkippedCount int
	ErrorCount   int
}

const JunitReportXml = `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="com.example.project.CalculatorTests" time="0.168" tests="6" errors="7" skipped="8" failures="9">
 <testcase name="addsTwoNumbers" classname="com.example.project.CalculatorTests" time="0.034"/>
 <testcase name="add(int, int, int)[1]" classname="com.example.project.CalculatorTests" time="0.022"/>
 <testcase name="add(int, int, int)[2]" classname="com.example.project.CalculatorTests" time="0.001"/>
 <testcase name="add(int, int, int)[3]" classname="com.example.project.CalculatorTests" time="0.002"/>
 <testcase name="add(int, int, int)[4]" classname="com.example.project.CalculatorTests" time="0.001"/>
</testsuite>`

// Mock function to parse JUnit XML
func parseJunitXML(xmlContent string) (MockJunitTestStats, error) {
	type JUnitTestSuite struct {
		XMLName  xml.Name `xml:"testsuite"`
		Tests    int      `xml:"tests,attr"`
		Failures int      `xml:"failures,attr"`
		Errors   int      `xml:"errors,attr"`
		Skipped  int      `xml:"skipped,attr"`
	}

	var parsed JUnitTestSuite
	err := xml.Unmarshal([]byte(xmlContent), &parsed)
	if err != nil {
		return MockJunitTestStats{}, fmt.Errorf("failed to parse JUnit XML: %w", err)
	}

	return MockJunitTestStats{
		TestCount:    parsed.Tests,
		FailCount:    parsed.Failures,
		PassCount:    max(0, parsed.Tests-(parsed.Failures+parsed.Errors+parsed.Skipped)),
		SkippedCount: parsed.Skipped,
		ErrorCount:   parsed.Errors,
	}, nil
}

func MockGetJunitDataMaps(pipelineId, buildNumber string, aggregateData MockJunitTestStats) (map[string]string, map[string]interface{}) {
	return map[string]string{
			"pipelineId": pipelineId,
			"buildId":    buildNumber,
		}, map[string]interface{}{
			"total_tests":   aggregateData.TestCount,
			"failed_tests":  aggregateData.FailCount,
			"passed_tests":  aggregateData.PassCount,
			"skipped_tests": aggregateData.SkippedCount,
			"errors_count":  aggregateData.ErrorCount,
		}
}

func TestJunitAggregator_Aggregate(t *testing.T) {
	junitReport, err := parseJunitXML(JunitReportXml)
	if err != nil {
		t.Fatalf("Error parsing JUnit XML: %v", err)
	}

	expectedStats := MockJunitTestStats{
		TestCount:    6,
		FailCount:    9,
		PassCount:    0,
		SkippedCount: 8,
		ErrorCount:   7,
	}

	if junitReport != expectedStats {
		t.Errorf("JUnit aggregation mismatch: got %+v, expected %+v", junitReport, expectedStats)
	}
}

func TestMockGetJunitDataMaps(t *testing.T) {
	pipelineId := "pipeline_001"
	buildNumber := "build_123"
	aggregateData := MockJunitTestStats{
		TestCount:    10,
		FailCount:    2,
		PassCount:    6,
		SkippedCount: 1,
		ErrorCount:   1,
	}

	tags, fields := MockGetJunitDataMaps(pipelineId, buildNumber, aggregateData)

	expectedTags := map[string]string{
		"pipelineId": pipelineId,
		"buildId":    buildNumber,
	}

	expectedFields := map[string]interface{}{
		"total_tests":   10,
		"failed_tests":  2,
		"passed_tests":  6,
		"skipped_tests": 1,
		"errors_count":  1,
	}

	for key, expectedValue := range expectedTags {
		if tags[key] != expectedValue {
			t.Errorf("Mismatch in tags: got %s = %v, expected %v", key, tags[key], expectedValue)
		}
	}

	for key, expectedValue := range expectedFields {
		if fields[key] != expectedValue {
			t.Errorf("Mismatch in fields: got %s = %v, expected %v", key, fields[key], expectedValue)
		}
	}
}

func MockJunitComputeBuildResultDifferences(currentValues, previousValues map[string]float64) string {
	var csvBuffer strings.Builder
	writer := csv.NewWriter(&csvBuffer)

	header := []string{"ResultType", "CurrentBuild", "PreviousBuild", "Difference", "PercentageDifference"}
	if err := writer.Write(header); err != nil {
		fmt.Println("Error writing CSV header:", err)
	}

	for key := range currentValues {
		currentValue := currentValues[key]
		previousValue := previousValues[key]

		diff := currentValue - previousValue
		percentageDiff := 0.0
		if previousValue != 0 {
			percentageDiff = (diff / previousValue) * 100
		}

		row := []string{
			key,
			fmt.Sprintf("%.2f", currentValue),
			fmt.Sprintf("%.2f", previousValue),
			fmt.Sprintf("%.2f", diff),
			fmt.Sprintf("%.2f%%", percentageDiff),
		}

		if err := writer.Write(row); err != nil {
			fmt.Println("Error writing CSV row:", err)
		}
	}

	writer.Flush()
	return csvBuffer.String()
}

func TestComputeJJunitBuildResultDifferences(t *testing.T) {
	currentValues := map[string]float64{
		"branch_covered_sum":      114,
		"branch_missed_sum":       2,
		"branch_total_sum":        116,
		"class_covered_sum":       10,
		"class_missed_sum":        0,
		"class_total_sum":         10,
		"complexity_covered_sum":  140,
		"complexity_missed_sum":   2,
		"complexity_total_sum":    142,
		"instruction_covered_sum": 1240,
		"instruction_missed_sum":  0,
		"instruction_total_sum":   1240,
		"line_covered_sum":        244,
		"line_missed_sum":         0,
		"line_total_sum":          244,
		"method_covered_sum":      84,
		"method_missed_sum":       0,
		"method_total_sum":        84,
	}

	previousValues := map[string]float64{
		"branch_covered_sum":      114,
		"branch_missed_sum":       2,
		"branch_total_sum":        116,
		"class_covered_sum":       10,
		"class_missed_sum":        0,
		"class_total_sum":         10,
		"complexity_covered_sum":  140,
		"complexity_missed_sum":   2,
		"complexity_total_sum":    142,
		"instruction_covered_sum": 1240,
		"instruction_missed_sum":  0,
		"instruction_total_sum":   1240,
		"line_covered_sum":        244,
		"line_missed_sum":         0,
		"line_total_sum":          244,
		"method_covered_sum":      84,
		"method_missed_sum":       0,
		"method_total_sum":        84,
	}

	resultStr := MockJunitComputeBuildResultDifferences(currentValues, previousValues)

	if !strings.Contains(resultStr, "ResultType,CurrentBuild,PreviousBuild,Difference,PercentageDifference") {
		t.Errorf("Expected header not found in result: %q", resultStr)
	}
}

package plugin

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
)

type MockTestNGReport struct {
	Ignored int
	Total   int
	Passed  int
	Failed  int
	Skipped int
}

const XmlTestNgReport = `<?xml version="1.0" encoding="UTF-8"?>
<testng-results ignored="0" total="10" passed="6" failed="2" skipped="2">
  <reporter-output>
  </reporter-output>
  <suite started-at="2025-01-22T22:46:54 IST" name="Suite 2" finished-at="2025-01-22T22:46:54 IST" duration-ms="4">
    <groups>
    </groups>
  </suite>
</testng-results>`

func parseTestNgXML(xmlContent string) (MockTestNGReport, error) {
	type TestNGResults struct {
		XMLName xml.Name `xml:"testng-results"`
		Ignored int      `xml:"ignored,attr"`
		Total   int      `xml:"total,attr"`
		Passed  int      `xml:"passed,attr"`
		Failed  int      `xml:"failed,attr"`
		Skipped int      `xml:"skipped,attr"`
	}

	var parsed TestNGResults
	err := xml.Unmarshal([]byte(xmlContent), &parsed)
	if err != nil {
		return MockTestNGReport{}, fmt.Errorf("failed to parse TestNG XML: %w", err)
	}

	return MockTestNGReport{
		Ignored: parsed.Ignored,
		Total:   parsed.Total,
		Passed:  parsed.Passed,
		Failed:  parsed.Failed,
		Skipped: parsed.Skipped,
	}, nil
}

func MockGetTestNgDataMaps(pipelineId, buildNumber string, aggregateData MockTestNGReport) (map[string]string, map[string]interface{}) {
	return map[string]string{
			"pipelineId": pipelineId,
			"buildId":    buildNumber,
		}, map[string]interface{}{
			"ignored": aggregateData.Ignored,
			"total":   aggregateData.Total,
			"passed":  aggregateData.Passed,
			"failed":  aggregateData.Failed,
			"skipped": aggregateData.Skipped,
		}
}

func TestXmlTestNgReport(t *testing.T) {
	testNgReport, err := parseTestNgXML(XmlTestNgReport)
	if err != nil {
		t.Fatalf("Error parsing TestNG XML: %v", err)
	}

	expectedStats := MockTestNGReport{
		Ignored: 0,
		Total:   10,
		Passed:  6,
		Failed:  2,
		Skipped: 2,
	}

	if testNgReport != expectedStats {
		t.Errorf("TestNG aggregation mismatch: got %+v, expected %+v", testNgReport, expectedStats)
	}
}

func TestMockGetTestNgDataMaps(t *testing.T) {
	pipelineId := "pipeline_001"
	buildNumber := "build_123"
	aggregateData := MockTestNGReport{
		Ignored: 2,
		Total:   10,
		Passed:  6,
		Failed:  2,
		Skipped: 2,
	}

	tags, fields := MockGetTestNgDataMaps(pipelineId, buildNumber, aggregateData)

	expectedTags := map[string]string{
		"pipelineId": pipelineId,
		"buildId":    buildNumber,
	}

	expectedFields := map[string]interface{}{
		"ignored": 2,
		"total":   10,
		"passed":  6,
		"failed":  2,
		"skipped": 2,
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

func MockTestNgComputeBuildResultDifferences(currentValues, previousValues map[string]float64) string {
	var csvBuffer strings.Builder
	header := "ResultType,CurrentBuild,PreviousBuild,Difference,PercentageDifference\n"
	csvBuffer.WriteString(header)

	for key := range currentValues {
		currentValue := currentValues[key]
		previousValue := previousValues[key]

		diff := currentValue - previousValue
		percentageDiff := mockTestNgComputePercentageDiff(previousValue, currentValue)

		row := fmt.Sprintf("%s,%.2f,%.2f,%.2f,%.2f%%\n", key, currentValue, previousValue, diff, percentageDiff)
		csvBuffer.WriteString(row)
	}

	return csvBuffer.String()
}

func TestComputeTestNgBuildResultDifferences(t *testing.T) {
	currentValues := map[string]float64{
		"ignored": 2,
		"total":   10,
		"passed":  6,
		"failed":  2,
		"skipped": 2,
	}

	previousValues := map[string]float64{
		"ignored": 1,
		"total":   10,
		"passed":  7,
		"failed":  2,
		"skipped": 1,
	}

	resultStr := MockTestNgComputeBuildResultDifferences(currentValues, previousValues)

	expectedCsvRows := []string{
		"ResultType,CurrentBuild,PreviousBuild,Difference,PercentageDifference",
		"total,10.00,10.00,0.00,0.00%",
		"passed,6.00,7.00,-1.00,-14.29%",
		"failed,2.00,2.00,0.00,0.00%",
		"skipped,2.00,1.00,1.00,100.00%",
		"ignored,2.00,1.00,1.00,100.00%",
	}

	for _, expectedRow := range expectedCsvRows {
		if !strings.Contains(resultStr, expectedRow) {
			t.Errorf("Expected row not found in result: %q", expectedRow)
		}
	}
}

func mockTestNgComputePercentageDiff(previous, current float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return ((current - previous) / math.Abs(previous)) * 100
}

func TestGetTestNgDataMaps(t *testing.T) {
	aggregateData := TestNGReport{
		AggregatedResults: Results{
			Total:      30,
			Failures:   6,
			Skipped:    3,
			DurationMS: 150.3,
		},
	}

	tags, fields := GetTestNgDataMaps("pipeline456", "build789", aggregateData)

	expectedTags := map[string]string{
		"pipelineId": "pipeline456",
		"buildId":    "build789",
	}

	expectedFields := map[string]interface{}{
		"total_cases":   30,
		"total_failed":  6,
		"total_skipped": 3,
		"duration_ms":   150.3,
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

func TestShowTestNgStats(t *testing.T) {
	tags := map[string]string{"pipelineId": "pipeline_2", "buildId": "build_456"}
	fields := map[string]interface{}{
		"total_cases":   60,
		"total_failed":  10,
		"total_skipped": 5,
		"duration_ms":   300.00,
	}

	// Capture stdout
	var output bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := ShowTestNgStats(tags, fields)

	w.Close()
	os.Stdout = oldStdout

	_, _ = output.ReadFrom(r)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedChecks := []string{
		"Total Cases", "60",
		"Total Failed", "10",
		"Total Skipped", "5",
		"Total Duration", "300.00",
	}

	for i := 0; i < len(expectedChecks); i += 2 {
		if !strings.Contains(output.String(), expectedChecks[i]) || !strings.Contains(output.String(), expectedChecks[i+1]) {
			t.Errorf("Expected output to contain both '%s' and '%s', but got:\n%s", expectedChecks[i], expectedChecks[i+1], output.String())
		}
	}
}

func TestMockComputeTestNgBuildResultDifferences(t *testing.T) {
	currentValues := map[string]float64{
		"total_cases":   70,
		"total_failed":  12,
		"total_skipped": 8,
		"duration_ms":   350.0,
	}

	previousValues := map[string]float64{
		"total_cases":   60,
		"total_failed":  10,
		"total_skipped": 5,
		"duration_ms":   300.0,
	}

	resultStr := MockTestNgComputeBuildResultDifferences(currentValues, previousValues)

	expectedCsvRows := []string{
		"ResultType,CurrentBuild,PreviousBuild,Difference,PercentageDifference",
		"total_cases,70.00,60.00,10.00,16.67%",
		"total_failed,12.00,10.00,2.00,20.00%",
		"total_skipped,8.00,5.00,3.00,60.00%",
		"duration_ms,350.00,300.00,50.00,16.67%",
	}

	for _, expectedRow := range expectedCsvRows {
		if !strings.Contains(resultStr, expectedRow) {
			t.Errorf("Expected row not found in result: %q", expectedRow)
		}
	}
}

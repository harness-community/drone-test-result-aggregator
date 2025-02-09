package plugin

import (
	"encoding/xml"
	"fmt"
	"math"
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

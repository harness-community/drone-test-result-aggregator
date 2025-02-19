package plugin

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

const NunitTestXml = `<?xml version="1.0" encoding="utf-8"?>
<TestRun id="3e5652b8-f41b-4e07-b3d2-5f28f4f6cc08" name="@ubuntu 2025-01-22 23:28:25" xmlns="http://microsoft.com/schemas/VisualStudio/TeamTest/2010">
  <Times creation="2025-01-22T23:28:25.8095145+05:30" queuing="2025-01-22T23:28:25.8095146+05:30" start="2025-01-22T23:28:25.1581777+05:30" finish="2025-01-22T23:28:25.8173908+05:30" />
  <TestSettings name="default" id="7f157ee1-4e4e-4e47-ae51-e2721f37e8f9">
    <Deployment runDeploymentRoot="_ubuntu_2025-01-22_23_28_25" />
  </TestSettings>
  <Results>
    <UnitTestResult executionId="104b04fb-6049-4b09-af5c-c76330cc6b21" testId="7c13ea3d-4237-3744-118f-4b8a47460eb9" testName="Test2" computerName="ubuntu" duration="00:00:00.0003390" startTime="2025-01-22T23:28:25.6785310+05:30" endTime="2025-01-22T23:28:25.6788696+05:30" testType="13cdc9d9-ddb5-4fa4-a97d-d965ccfc6d4b" outcome="Passed" testListId="8c84fa94-04c1-424b-9868-57a2d4851a1d" relativeResultsDirectory="104b04fb-6049-4b09-af5c-c76330cc6b21">
      <Output>
        <StdOut>Suite2 - Test2 passed.</StdOut>
      </Output>
    </UnitTestResult>
    <UnitTestResult executionId="e4854018-8d7c-4df8-a7ae-3bb0175b217c" testId="38ee8d2c-88e3-5727-2cbd-76b175ac95c2" testName="Test1" computerName="ubuntu" duration="00:00:00.0166410" startTime="2025-01-22T23:28:25.6618423+05:30" endTime="2025-01-22T23:28:25.6784828+05:30" testType="13cdc9d9-ddb5-4fa4-a97d-d965ccfc6d4b" outcome="Passed" testListId="8c84fa94-04c1-424b-9868-57a2d4851a1d" relativeResultsDirectory="e4854018-8d7c-4df8-a7ae-3bb0175b217c">
      <Output>
        <StdOut>Suite2 - Test1 passed.</StdOut>
      </Output>
    </UnitTestResult>
    <UnitTestResult executionId="37b76398-bbb7-400b-b492-306c2fb5183f" testId="f9a07328-18a4-0d0d-a9ed-1e4e74e7218b" testName="IgnoredTest" computerName="ubuntu" duration="00:00:00.0015839" startTime="2025-01-22T23:28:25.6586400+05:30" endTime="2025-01-22T23:28:25.6601781+05:30" testType="13cdc9d9-ddb5-4fa4-a97d-d965ccfc6d4b" outcome="NotExecuted" testListId="8c84fa94-04c1-424b-9868-57a2d4851a1d" relativeResultsDirectory="37b76398-bbb7-400b-b492-306c2fb5183f">
      <Output>
        <StdOut>This test is ignored for demonstration.</StdOut>
        <ErrorInfo>
          <Message>This test is ignored for demonstration.</Message>
        </ErrorInfo>
      </Output>
    </UnitTestResult>
  </Results>
  <ResultSummary outcome="Completed">
    <Counters total="3" executed="2" passed="2" failed="0" error="0" />
  </ResultSummary>
</TestRun>`

type MockNunitTestStats struct {
	Total        int
	Executed     int
	Passed       int
	Failed       int
	Error        int
	Timeout      int
	Aborted      int
	Inconclusive int
	NotRunnable  int
	NotExecuted  int
	Disconnected int
	Warning      int
	Completed    int
	InProgress   int
	Pending      int
}

func parseNunitXML(xmlContent string) (MockNunitTestStats, error) {
	type NunitSummary struct {
		XMLName  xml.Name `xml:"TestRun"`
		Counters struct {
			Total        int `xml:"total,attr"`
			Executed     int `xml:"executed,attr"`
			Passed       int `xml:"passed,attr"`
			Failed       int `xml:"failed,attr"`
			Error        int `xml:"error,attr"`
			Timeout      int `xml:"timeout,attr"`
			Aborted      int `xml:"aborted,attr"`
			Inconclusive int `xml:"inconclusive,attr"`
			NotRunnable  int `xml:"notRunnable,attr"`
			NotExecuted  int `xml:"notExecuted,attr"`
			Disconnected int `xml:"disconnected,attr"`
			Warning      int `xml:"warning,attr"`
			Completed    int `xml:"completed,attr"`
			InProgress   int `xml:"inProgress,attr"`
			Pending      int `xml:"pending,attr"`
		} `xml:"ResultSummary>Counters"`
	}

	var parsed NunitSummary
	err := xml.Unmarshal([]byte(xmlContent), &parsed)
	if err != nil {
		return MockNunitTestStats{}, fmt.Errorf("failed to parse NUnit XML: %w", err)
	}

	return MockNunitTestStats{
		Total:        parsed.Counters.Total,
		Executed:     parsed.Counters.Executed,
		Passed:       parsed.Counters.Passed,
		Failed:       parsed.Counters.Failed,
		Error:        parsed.Counters.Error,
		Timeout:      parsed.Counters.Timeout,
		Aborted:      parsed.Counters.Aborted,
		Inconclusive: parsed.Counters.Inconclusive,
		NotRunnable:  parsed.Counters.NotRunnable,
		NotExecuted:  parsed.Counters.NotExecuted,
		Disconnected: parsed.Counters.Disconnected,
		Warning:      parsed.Counters.Warning,
		Completed:    parsed.Counters.Completed,
		InProgress:   parsed.Counters.InProgress,
		Pending:      parsed.Counters.Pending,
	}, nil
}

func TestNunitAggregator(t *testing.T) {
	nunitReport, err := parseNunitXML(NunitTestXml)
	if err != nil {
		t.Fatalf("Error parsing NUnit XML: %v", err)
	}

	expectedStats := MockNunitTestStats{
		Total:    3,
		Executed: 2,
		Passed:   2,
		Failed:   0,
	}

	if nunitReport != expectedStats {
		t.Errorf("NUnit aggregation mismatch: got %+v, expected %+v", nunitReport, expectedStats)
	}
}

func TestComputeNunitBuildResultDifferences(t *testing.T) {
	currentValues := map[string]float64{
		"total": 3, "executed": 2, "passed": 2, "failed": 0,
	}
	previousValues := map[string]float64{
		"total": 3, "executed": 2, "passed": 2, "failed": 0,
	}

	result := MockComputeNunitBuildResultDifferences(currentValues, previousValues)
	expectedCsvRows := []string{
		"ResultType, CurrentBuild, PreviousBuild, Difference, PercentageDifference",
		"total, 3.00, 3.00, 0.00, 0.00",
		"executed, 2.00, 2.00, 0.00, 0.00",
		"passed, 2.00, 2.00, 0.00, 0.00",
		"failed, 0.00, 0.00, 0.00, 0.00",
	}

	for _, expectedRow := range expectedCsvRows {
		if !strings.Contains(result, expectedRow) {
			t.Errorf("Expected row not found in result: %q", expectedRow)
		}
	}
}

func MockComputeNunitBuildResultDifferences(currentValues, previousValues map[string]float64) string {
	var csvBuffer strings.Builder
	csvBuffer.WriteString("ResultType, CurrentBuild, PreviousBuild, Difference, PercentageDifference\n")

	for key := range currentValues {
		currentValue := currentValues[key]
		previousValue := previousValues[key]
		diff := currentValue - previousValue
		percentageDiff := 0.0
		if previousValue != 0 {
			percentageDiff = (diff / previousValue) * 100
		}
		csvBuffer.WriteString(fmt.Sprintf("%s, %.2f, %.2f, %.2f, %.2f\n", key, currentValue, previousValue, diff, percentageDiff))
	}

	return csvBuffer.String()
}

func TestCalculateNunitAggregate(t *testing.T) {
	reports := []TestRunSummary{
		{TotalCases: 5, TotalPassed: 3, TotalFailed: 1, TotalSkipped: 1},
		{TotalCases: 8, TotalPassed: 6, TotalFailed: 2, TotalSkipped: 0},
	}

	result := CalculateNunitAggregate(reports)

	if result.TotalCases != 13 {
		t.Errorf("Expected TotalCases = 13, got %d", result.TotalCases)
	}
	if result.TotalPassed != 9 {
		t.Errorf("Expected TotalPassed = 9, got %d", result.TotalPassed)
	}
	if result.TotalFailed != 3 {
		t.Errorf("Expected TotalFailed = 3, got %d", result.TotalFailed)
	}
	if result.TotalSkipped != 1 {
		t.Errorf("Expected TotalSkipped = 1, got %d", result.TotalSkipped)
	}
}

func TestGetNunitDataMaps(t *testing.T) {
	aggregateData := TestRunSummary{
		TotalCases:   10,
		TotalPassed:  7,
		TotalFailed:  2,
		TotalSkipped: 1,
	}

	tags, fields := GetNunitDataMaps("pipeline123", "build45", aggregateData)

	expectedTags := map[string]string{
		"pipelineId": "pipeline123",
		"buildId":    "build45",
	}

	expectedFields := map[string]interface{}{
		"total_cases":   10,
		"total_passed":  7,
		"total_failed":  2,
		"total_skipped": 1,
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

func TestShowNunitStats(t *testing.T) {
	tags := map[string]string{"pipelineId": "pipeline_1", "buildId": "build_123"}
	fields := map[string]interface{}{
		"total_cases":   100,
		"total_passed":  80,
		"total_failed":  10,
		"total_skipped": 10,
	}

	var output bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := ShowNunitStats(tags, fields)

	w.Close()
	os.Stdout = oldStdout

	_, _ = output.ReadFrom(r)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedChecks := []string{
		"Total Cases", "100",
		"Total Passed", "80",
		"Total Failed", "10",
		"Total Skipped", "10",
	}

	for i := 0; i < len(expectedChecks); i += 2 {
		if !strings.Contains(output.String(), expectedChecks[i]) || !strings.Contains(output.String(), expectedChecks[i+1]) {
			t.Errorf("Expected output to contain both '%s' and '%s', but got:\n%s", expectedChecks[i], expectedChecks[i+1], output.String())
		}
	}
}

package plugin

import (
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

type MockJacocoTestStats struct {
	InstructionCoveredSum float64
	InstructionMissedSum  float64
	BranchCoveredSum      float64
	BranchMissedSum       float64
	LineCoveredSum        float64
	LineMissedSum         float64
	ComplexityCoveredSum  float64
	ComplexityMissedSum   float64
	MethodCoveredSum      float64
	MethodMissedSum       float64
	ClassCoveredSum       float64
	ClassMissedSum        float64
}

const JacocoReportXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><!DOCTYPE report PUBLIC "-//JACOCO//DTD Report 1.1//EN"
        "report.dtd">
<report name="JaCoCo Coverage Report">
    <sessioninfo id="ubuntu-20dd7d00" start="1728987209281" dump="1728987210434"/>
    <sessioninfo id="ubuntu-c45dc566" start="1728987219096" dump="1728987220330"/>
    <counter type="INSTRUCTION" missed="0" covered="620"/>
    <counter type="BRANCH" missed="1" covered="57"/>
    <counter type="LINE" missed="0" covered="122"/>
    <counter type="COMPLEXITY" missed="1" covered="70"/>
    <counter type="METHOD" missed="0" covered="42"/>
    <counter type="CLASS" missed="0" covered="5"/>
</report>`

func parseJacocoXML(xmlContent string) (MockJacocoTestStats, error) {
	type Counter struct {
		Type    string `xml:"type,attr"`
		Missed  int    `xml:"missed,attr"`
		Covered int    `xml:"covered,attr"`
	}

	type Report struct {
		XMLName  xml.Name  `xml:"report"`
		Counters []Counter `xml:"counter"`
	}

	var parsed Report
	err := xml.Unmarshal([]byte(xmlContent), &parsed)
	if err != nil {
		return MockJacocoTestStats{}, fmt.Errorf("failed to parse Jacoco XML: %w", err)
	}

	var stats MockJacocoTestStats
	for _, counter := range parsed.Counters {
		switch counter.Type {
		case "INSTRUCTION":
			stats.InstructionCoveredSum = float64(counter.Covered)
			stats.InstructionMissedSum = float64(counter.Missed)
		case "BRANCH":
			stats.BranchCoveredSum = float64(counter.Covered)
			stats.BranchMissedSum = float64(counter.Missed)
		case "LINE":
			stats.LineCoveredSum = float64(counter.Covered)
			stats.LineMissedSum = float64(counter.Missed)
		case "COMPLEXITY":
			stats.ComplexityCoveredSum = float64(counter.Covered)
			stats.ComplexityMissedSum = float64(counter.Missed)
		case "METHOD":
			stats.MethodCoveredSum = float64(counter.Covered)
			stats.MethodMissedSum = float64(counter.Missed)
		case "CLASS":
			stats.ClassCoveredSum = float64(counter.Covered)
			stats.ClassMissedSum = float64(counter.Missed)
		}
	}

	return stats, nil
}

func TestJacocoAggregator(t *testing.T) {
	jacocoReport, err := parseJacocoXML(JacocoReportXml)
	if err != nil {
		t.Fatalf("Error parsing Jacoco XML: %v", err)
	}

	expectedStats := MockJacocoTestStats{
		InstructionCoveredSum: 620,
		InstructionMissedSum:  0,
		BranchCoveredSum:      57,
		BranchMissedSum:       1,
		LineCoveredSum:        122,
		LineMissedSum:         0,
		ComplexityCoveredSum:  70,
		ComplexityMissedSum:   1,
		MethodCoveredSum:      42,
		MethodMissedSum:       0,
		ClassCoveredSum:       5,
		ClassMissedSum:        0,
	}

	if jacocoReport != expectedStats {
		t.Errorf("Jacoco aggregation mismatch: got %+v, expected %+v", jacocoReport, expectedStats)
	}
}

func MockJacocoComputeBuildResultDifferences(currentValues, previousValues map[string]float64) string {
	var csvBuffer strings.Builder
	writer := csv.NewWriter(&csvBuffer)

	header := []string{"ResultType", "CurrentBuild", "PreviousBuild", "Difference", "PercentageDifference"}
	_ = writer.Write(header)

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
		_ = writer.Write(row)
	}

	writer.Flush()
	return csvBuffer.String()
}

func TestComputeJacocoBuildResultDifferences(t *testing.T) {
	currentValues := map[string]float64{
		"branch_covered_sum":     57,
		"branch_missed_sum":      1,
		"line_covered_sum":       122,
		"line_missed_sum":        0,
		"complexity_covered_sum": 70,
		"complexity_missed_sum":  1,
		"method_covered_sum":     42,
		"method_missed_sum":      0,
		"class_covered_sum":      5,
		"class_missed_sum":       0,
	}

	previousValues := map[string]float64{
		"branch_covered_sum":     55,
		"branch_missed_sum":      2,
		"line_covered_sum":       120,
		"line_missed_sum":        1,
		"complexity_covered_sum": 65,
		"complexity_missed_sum":  2,
		"method_covered_sum":     40,
		"method_missed_sum":      1,
		"class_covered_sum":      5,
		"class_missed_sum":       0,
	}

	resultStr := MockJacocoComputeBuildResultDifferences(currentValues, previousValues)

	expectedCsvRows := []string{
		"ResultType,CurrentBuild,PreviousBuild,Difference,PercentageDifference",
		"branch_covered_sum,57.00,55.00,2.00,3.64%",
		"branch_missed_sum,1.00,2.00,-1.00,-50.00%",
		"line_covered_sum,122.00,120.00,2.00,1.67%",
		"line_missed_sum,0.00,1.00,-1.00,-100.00%",
		"complexity_covered_sum,70.00,65.00,5.00,7.69%",
	}

	for _, expectedRow := range expectedCsvRows {
		if !strings.Contains(resultStr, expectedRow) {
			t.Errorf("Expected row not found in result: %q", expectedRow)
		}
	}
}

func TestCalculateJacocoAggregate(t *testing.T) {
	reports := []Report{
		{
			Counters: []Counter{
				{Type: "INSTRUCTION", Covered: 10, Missed: 5},
				{Type: "BRANCH", Covered: 8, Missed: 2},
			},
		},
		{
			Counters: []Counter{
				{Type: "INSTRUCTION", Covered: 15, Missed: 10},
				{Type: "BRANCH", Covered: 6, Missed: 4},
			},
		},
	}

	result := CalculateJacocoAggregate(reports)

	if result.InstructionTotalSum != 40 {
		t.Errorf("Expected InstructionTotalSum to be 40, got %f", result.InstructionTotalSum)
	}
	if result.InstructionCoveredSum != 25 {
		t.Errorf("Expected InstructionCoveredSum to be 25, got %f", result.InstructionCoveredSum)
	}
	if result.InstructionMissedSum != 15 {
		t.Errorf("Expected InstructionMissedSum to be 15, got %f", result.InstructionMissedSum)
	}
	if result.BranchTotalSum != 20 {
		t.Errorf("Expected BranchTotalSum to be 20, got %f", result.BranchTotalSum)
	}
}

func TestCalculatePercentage(t *testing.T) {
	tests := []struct {
		covered, missed int
		expected        float64
	}{
		{10, 10, 50.0},
		{0, 10, 0.0},
		{10, 0, 100.0},
		{0, 0, 0.0},
	}

	for _, test := range tests {
		result := CalculatePercentage(test.covered, test.missed)
		if result != test.expected {
			t.Errorf("CalculatePercentage(%d, %d) = %f; want %f", test.covered, test.missed, result, test.expected)
		}
	}
}
func TestGetJacocoDataMaps(t *testing.T) {
	report := Report{
		JacocoAggregateData: JacocoAggregateData{ // ✅ Initialize the embedded struct properly
			InstructionTotalSum:   50,
			InstructionCoveredSum: 30,
			InstructionMissedSum:  20,
			BranchTotalSum:        40,
			BranchCoveredSum:      25,
			BranchMissedSum:       15,
		},
	}

	tags, fields := GetJacocoDataMaps("pipeline123", "build45", report)

	if tags["pipelineId"] != "pipeline123" {
		t.Errorf("Expected pipelineId to be 'pipeline123', got %s", tags["pipelineId"])
	}
	if tags["buildId"] != "build45" {
		t.Errorf("Expected buildId to be 'build45', got %s", tags["buildId"])
	}
	if fields["instruction_total_sum"].(float64) != 50 {
		t.Errorf("Expected instruction_total_sum to be 50, got %f", fields["instruction_total_sum"].(float64))
	}
	if fields["branch_total_sum"].(float64) != 40 {
		t.Errorf("Expected branch_total_sum to be 40, got %f", fields["branch_total_sum"].(float64))
	}
}

func TestShowJacocoStats(t *testing.T) {
	tags := map[string]string{"pipelineId": "pipe_1", "buildId": "123"}
	fields := map[string]interface{}{
		"instruction_total_sum":   100.0,
		"instruction_covered_sum": 80.0,
		"instruction_missed_sum":  20.0,
	}

	err := ShowJacocoStats(tags, fields)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func MockGetPreviousBuildId(buildIds []string, currentBuildId string) (int, error) {
	currentBuild, err := strconv.Atoi(currentBuildId)
	if err != nil {
		return 0, fmt.Errorf("invalid currentBuildId: %s", currentBuildId)
	}

	var builds []int
	for _, b := range buildIds {
		num, err := strconv.Atoi(b)
		if err == nil {
			builds = append(builds, num)
		}
	}

	var prevBuild int
	for _, id := range builds {
		if id < currentBuild && id > prevBuild {
			prevBuild = id
		}
	}

	if prevBuild == 0 {
		return 0, errors.New("no previous build ID found")
	}

	return prevBuild, nil
}

func TestMockGetPreviousBuildId(t *testing.T) {
	tests := []struct {
		name          string
		buildIds      []string
		currentBuild  string
		expectedBuild int
		expectErr     bool
	}{
		{"Valid previous build", []string{"105", "102", "100"}, "105", 102, false},
		{"No previous build", []string{"100"}, "100", 0, true},
		{"Invalid build in list", []string{"abc", "99"}, "100", 99, false},
		{"Invalid currentBuildId", []string{"100", "95"}, "xyz", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevBuild, err := MockGetPreviousBuildId(tt.buildIds, tt.currentBuild)

			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}

			if prevBuild != tt.expectedBuild {
				t.Errorf("Expected previous build ID: %d, got: %d", tt.expectedBuild, prevBuild)
			}
		})
	}
}

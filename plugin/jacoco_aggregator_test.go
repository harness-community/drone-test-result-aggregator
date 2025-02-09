package plugin

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
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

package plugin

import (
	"fmt"
	"strings"
	"testing"
)

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

func TestJacocoAggregator(t *testing.T) {
	tagsMap, fieldsMap := MockAggregate[Report](JacocoReportXml, CalculateJacocoAggregate, GetJacocoDataMaps)

	expectedTagsMap := map[string]string{
		"buildId":    mockBuildNumber,
		"pipelineId": mockPipelineId,
	}

	expectedFieldsMap := map[string]interface{}{
		"branch_covered_sum":      57,
		"branch_missed_sum":       1,
		"branch_total_sum":        58,
		"class_covered_sum":       5,
		"class_missed_sum":        0,
		"class_total_sum":         5,
		"complexity_covered_sum":  70,
		"complexity_missed_sum":   1,
		"complexity_total_sum":    71,
		"instruction_covered_sum": 620,
		"instruction_missed_sum":  0,
		"instruction_total_sum":   620,
		"line_covered_sum":        122,
		"line_missed_sum":         0,
		"line_total_sum":          122,
		"method_covered_sum":      42,
		"method_missed_sum":       0,
		"method_total_sum":        42,
	}

	for k := range expectedTagsMap {
		if tagsMap[k] != expectedTagsMap[k] {
			t.Errorf("Mismatch in TagsMap for key %q: got %v, expected %v", k, tagsMap[k], expectedTagsMap[k])
		}
	}

	for k := range expectedFieldsMap {
		gotVal := fmt.Sprintf("%v", fieldsMap[k])
		expectedVal := fmt.Sprintf("%v", expectedFieldsMap[k])
		if gotVal != expectedVal {
			t.Errorf("Mismatch in FieldsMap for key %q: got %v, expected %v", k, gotVal, expectedVal)
		}
	}
}

func TestComputeJacocoBuildResultDifferences(t *testing.T) {

	currentValues := map[string]float64{
		"instruction_total_sum":   1240,
		"instruction_covered_sum": 1240,
		"instruction_missed_sum":  0,
		"branch_total_sum":        116,
		"branch_covered_sum":      114,
		"branch_missed_sum":       2,
		"line_total_sum":          244,
		"line_covered_sum":        244,
		"line_missed_sum":         0,
		"complexity_total_sum":    142,
		"complexity_covered_sum":  140,
		"complexity_missed_sum":   2,
		"method_total_sum":        84,
		"method_covered_sum":      84,
		"method_missed_sum":       0,
		"class_total_sum":         10,
		"class_covered_sum":       10,
		"class_missed_sum":        0,
	}

	previousValues := map[string]float64{
		"instruction_total_sum":   1240,
		"instruction_covered_sum": 1240,
		"instruction_missed_sum":  0,
		"branch_total_sum":        116,
		"branch_covered_sum":      114,
		"branch_missed_sum":       2,
		"line_total_sum":          244,
		"line_covered_sum":        244,
		"line_missed_sum":         0,
		"complexity_total_sum":    142,
		"complexity_covered_sum":  140,
		"complexity_missed_sum":   2,
		"method_total_sum":        84,
		"method_covered_sum":      84,
		"method_missed_sum":       0,
		"class_total_sum":         10,
		"class_covered_sum":       10,
		"class_missed_sum":        0,
	}

	expectedResultDiffs := []ResultDiff{}
	for field, currValue := range currentValues {
		prevValue := previousValues[field]

		expectedResultDiffs = append(expectedResultDiffs, ResultDiff{
			FieldName:            field,
			CurrentBuildValue:    currValue,
			PreviousBuildValue:   prevValue,
			Difference:           currValue - prevValue,
			PercentageDifference: 0, // Since values are identical, percentage difference is 0
			IsCompareValid:       true,
		})
	}

	result := ComputeBuildResultDifferences(currentValues, previousValues)
	expectedCsvRows := []string{
		"ResultType, CurrentBuild, PreviousBuild, Difference, PercentageDifference",
		"instruction_missed_sum, 0.00, 0.00, 0.00, 0.00",
		"method_missed_sum, 0.00, 0.00, 0.00, 0.00",
		"class_covered_sum, 10.00, 10.00, 0.00, 0.00",
		"line_total_sum, 244.00, 244.00, 0.00, 0.00",
		"complexity_covered_sum, 140.00, 140.00, 0.00, 0.00",
		"instruction_total_sum, 1240.00, 1240.00, 0.00, 0.00",
		"class_missed_sum, 0.00, 0.00, 0.00, 0.00",
		"instruction_covered_sum, 1240.00, 1240.00, 0.00, 0.00",
		"line_missed_sum, 0.00, 0.00, 0.00, 0.00",
		"branch_covered_sum, 114.00, 114.00, 0.00, 0.00",
		"branch_missed_sum, 2.00, 2.00, 0.00, 0.00",
		"method_total_sum, 84.00, 84.00, 0.00, 0.00",
		"line_covered_sum, 244.00, 244.00, 0.00, 0.00",
		"method_covered_sum, 84.00, 84.00, 0.00, 0.00",
		"branch_total_sum, 116.00, 116.00, 0.00, 0.00",
		"complexity_total_sum, 142.00, 142.00, 0.00, 0.00",
		"complexity_missed_sum, 2.00, 2.00, 0.00, 0.00",
		"class_total_sum, 10.00, 10.00, 0.00, 0.00",
	}

	for _, expectedRow := range expectedCsvRows {
		found := strings.Contains(result, expectedRow)
		if !found {
			t.Errorf("Expected row not found in result: %q", expectedRow)
		}
	}
}

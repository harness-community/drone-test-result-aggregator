package plugin

import (
	"fmt"
	"math"
	"sort"
	"testing"
)

func TestXmlTestNgReport(t *testing.T) {
	tagsMap, fieldsMap := MockAggregate[TestNGResults](XmlTestNgReport,
		CalculateTestNgAggregate, GetTestNgDataMaps)
	expectedTagsMap := map[string]string{
		"buildId":    mockBuildNumber,
		"pipelineId": mockPipelineId,
	}

	expectedFieldsMap := map[string]interface{}{
		"ignored": 0,
		"total":   10,
		"passed":  6,
		"failed":  2,
		"skipped": 2,
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

const XmlTestNgReport = `<?xml version="1.0" encoding="UTF-8"?>
<testng-results ignored="0" total="10" passed="6" failed="2" skipped="2">
  <reporter-output>
  </reporter-output>
  <suite started-at="2025-01-22T22:46:54 IST" name="Suite 2" finished-at="2025-01-22T22:46:54 IST" duration-ms="4">
    <groups>
    </groups>
  </suite> <!-- Suite 1 -->
  <suite started-at="2025-01-22T22:46:54 IST" name="All Suites" finished-at="2025-01-22T22:46:54 IST" duration-ms="0">
    <groups>
    </groups>
  </suite> <!-- All Suites -->
</testng-results>`

func TestComputeTestNgBuildResultDifferences(t *testing.T) {
	currentBuildId := "5"
	previousBuildId := "4"
	pipelineId := "test_pipeline"
	groupId := "test_group"

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

	expectedResultDiffs := []ResultDiff{}
	for field, currValue := range currentValues {
		prevValue := previousValues[field]

		expectedResultDiffs = append(expectedResultDiffs, ResultDiff{
			FieldName:            field,
			CurrentBuildValue:    currValue,
			PreviousBuildValue:   prevValue,
			Difference:           currValue - prevValue,
			PercentageDifference: computePercentageDiff(prevValue, currValue),
			IsCompareValid:       true,
		})
	}

	result := ComputeBuildResultDifferences(currentBuildId, previousBuildId, pipelineId, groupId, currentValues, previousValues)

	resultDiffs, ok := result["result_differences"].([]ResultDiff)
	if !ok {
		t.Fatalf("Expected []ResultDiff, got %T", result["result_differences"])
	}

	sort.Slice(expectedResultDiffs, func(i, j int) bool {
		return expectedResultDiffs[i].FieldName < expectedResultDiffs[j].FieldName
	})
	sort.Slice(resultDiffs, func(i, j int) bool {
		return resultDiffs[i].FieldName < resultDiffs[j].FieldName
	})

	if len(resultDiffs) != len(expectedResultDiffs) {
		t.Fatalf("Expected %d results, got %d", len(expectedResultDiffs), len(resultDiffs))
	}

	for i, expected := range expectedResultDiffs {
		actual := resultDiffs[i]
		if actual.FieldName != expected.FieldName ||
			actual.CurrentBuildValue != expected.CurrentBuildValue ||
			actual.PreviousBuildValue != expected.PreviousBuildValue ||
			actual.Difference != expected.Difference ||
			math.Abs(actual.PercentageDifference-expected.PercentageDifference) > 0.0001 ||
			actual.IsCompareValid != expected.IsCompareValid {
			t.Errorf("Mismatch at index %d: got %+v, expected %+v", i, actual, expected)
		}
	}
}

func computePercentageDiff(previous, current float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return ((current - previous) / math.Abs(previous)) * 100
}

package plugin

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

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

	result := ComputeBuildResultDifferences(currentValues, previousValues)
	expectedCsvRows := []string{
		"ResultType, CurrentBuild, PreviousBuild, Difference, PercentageDifference",
		"total, 10.00, 10.00, 0.00, 0.00",
		"passed, 6.00, 7.00, -1.00, -14.29",
		"failed, 2.00, 2.00, 0.00, 0.00",
		"skipped, 2.00, 1.00, 1.00, 100.00",
		"ignored, 2.00, 1.00, 1.00, 100.00",
	}

	for _, expectedRow := range expectedCsvRows {
		found := strings.Contains(result, expectedRow)
		if !found {
			t.Errorf("Expected row not found in result: %q", expectedRow)
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

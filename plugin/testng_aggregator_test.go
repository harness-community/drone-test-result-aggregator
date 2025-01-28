package plugin

import (
	"fmt"
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

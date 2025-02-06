package plugin

import (
	"fmt"
	"strings"
	"testing"
)

const JunitReportXml = `<?xml version="1.0" encoding="UTF-8"?>
<testsuite xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="https://maven.apache.org/surefire/maven-surefire-plugin/xsd/surefire-test-report.xsd" version="3.0.2" name="com.example.project.CalculatorTests" time="0.168" tests="6" errors="7" skipped="8" failures="9">
 <properties>
   <property name="os.name" value="Linux"/>
   <property name="user.name" value="hns"/>
   <property name="path.separator" value=":"/>
   <property name="sun.io.unicode.encoding" value="UnicodeLittle"/>
   <property name="java.class.version" value="52.0"/>
 </properties>
 <testcase name="addsTwoNumbers" classname="com.example.project.CalculatorTests" time="0.034"/>
 <testcase name="add(int, int, int)[1]" classname="com.example.project.CalculatorTests" time="0.022"/>
 <testcase name="add(int, int, int)[2]" classname="com.example.project.CalculatorTests" time="0.001"/>
 <testcase name="add(int, int, int)[3]" classname="com.example.project.CalculatorTests" time="0.002"/>
 <testcase name="add(int, int, int)[4]" classname="com.example.project.CalculatorTests" time="0.001"/>
</testsuite>`

func TestJunitAggregator_Aggregate(t *testing.T) {

	tagsMap, fieldsMap := MockAggregate[JunitAggregatorData](JunitReportXml,
		CalculateJunitAggregate, GetJunitDataMaps)

	expectedTagsMap := map[string]string{
		"build_id":    mockBuildNumber,
		"pipeline_id": mockPipelineId,
	}

	expectedFieldsMap := map[string]interface{}{
		"errors":   7,
		"failures": 9,
		"skipped":  8,
		"tests":    6,
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

func TestComputeJunitBuildResultDifferences(t *testing.T) {
	currentValues := map[string]float64{
		"errors":   7,
		"failures": 9,
		"skipped":  8,
		"tests":    6,
	}

	previousValues := map[string]float64{
		"errors":   7,
		"failures": 9,
		"skipped":  8,
		"tests":    6,
	}

	resultStr := ComputeBuildResultDifferences(currentValues, previousValues)

	expectedCsvRows := []string{
		"ResultType, CurrentBuild, PreviousBuild, Difference, PercentageDifference",
		"errors, 7.00, 7.00, 0.00, 0.00",
		"failures, 9.00, 9.00, 0.00, 0.00",
		"skipped, 8.00, 8.00, 0.00, 0.00",
		"tests, 6.00, 6.00, 0.00, 0.00",
	}

	for _, expectedRow := range expectedCsvRows {
		found := strings.Contains(resultStr, expectedRow)
		if !found {
			t.Errorf("Expected row not found in result: %q", expectedRow)
		}
	}
}
